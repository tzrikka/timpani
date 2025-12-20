package github

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"

	"github.com/tzrikka/timpani/pkg/http/client"
)

const (
	accept = "application/vnd.github+json"
)

func (a *API) httpRequestPrep(ctx context.Context, path string) (l log.Logger, apiURL, token string, err error) {
	l = activity.GetLogger(ctx)

	var secrets map[string]string
	secrets, err = a.thrippy.LinkCreds(ctx)
	if err != nil {
		return l, "", "", err
	}

	baseURL := secrets["api_base_url"] // Added automatically for "github-app-jwt" links.
	if baseURL == "" {
		baseURL = secrets["base_url"] // Manual and optional for "github-app-user" and "github-user-pat" links.
		if baseURL == "" {
			baseURL = "https://api.github.com"
		} else {
			baseURL += "/api/v3" // GitHub Enterprise Server (GHES).
		}
	}

	apiURL, err = url.JoinPath(baseURL, path)
	if err != nil {
		l.Error("failed to construct GitHub API URL", slog.Any("error", err),
			slog.String("base_url", baseURL), slog.String("path", path))
		err = temporal.NewNonRetryableApplicationError(err.Error(), fmt.Sprintf("%T", err), err)
		return l, "", "", err
	}

	// "access_token" has a value only in "github-app-user" link secrets.
	// "pat" has a value only in "github-user-pat" link secrets.
	token = secrets["access_token"] + secrets["pat"]
	if token == "" {
		// JWTs generation is supported only for "github-app-jwt" links.
		token, err = generateJWT(secrets["client_id"], secrets["private_key"])
		if err != nil {
			msg := "failed to generate JWT for GitHub API call"
			l.Warn(msg, slog.Any("error", err), slog.String("link_id", a.thrippy.LinkID))
			err = temporal.NewNonRetryableApplicationError(msg, "error", err, a.thrippy.LinkID)
			return l, apiURL, token, err
		}
	}

	return l, apiURL, token, nil
}

// generateJWT generates a JSON Web Token (JWT) for a GitHub app. Based on:
// https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/generating-a-json-web-token-jwt-for-a-github-app
func generateJWT(clientID, privateKey string) (string, error) {
	// Input sanity checks.
	if clientID == "" {
		return "", errors.New("missing credential: client_id")
	}
	if privateKey == "" {
		return "", errors.New("missing credential: private_key")
	}

	// Parse the private key.
	privateKey = strings.ReplaceAll(privateKey, "\\n", "\n")
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return "", errors.New("failed to decode PEM private key")
	}

	pk, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	// Generate and sign the JWT.
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": now.Unix(),
		"exp": now.Add(10 * time.Minute).Unix(),
		"iss": clientID,
	})

	signedToken, err := token.SignedString(pk)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return signedToken, nil
}

// httpGet is a GitHub-specific HTTP GET wrapper for [client.HTTPRequest].
func (a *API) httpGet(ctx context.Context, path string, query url.Values, jsonResp any) error {
	l, apiURL, token, err := a.httpRequestPrep(ctx, path)
	if err != nil {
		return err
	}

	resp, _, err := client.HTTPRequest(ctx, http.MethodGet, apiURL, token, accept, "", query)
	if err != nil {
		l.Error("HTTP GET request error", slog.Any("error", err), slog.String("url", apiURL))
		return err
	}

	if err := json.Unmarshal(resp, jsonResp); err != nil {
		msg := "failed to decode HTTP response's JSON body"
		l.Error(msg, slog.Any("error", err), slog.String("url", apiURL))
		msg = fmt.Sprintf("%s: %v", msg, err)
		return temporal.NewNonRetryableApplicationError(msg, fmt.Sprintf("%T", err), err, apiURL)
	}

	l.Info("sent HTTP GET request", slog.String("link_id", a.thrippy.LinkID), slog.String("url", apiURL))
	return nil
}
