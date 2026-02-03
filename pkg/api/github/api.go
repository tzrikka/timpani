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
	defaultAccept = "application/vnd.github+json"
)

type tokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// httpDelete is a GitHub-specific HTTP DELETE wrapper for [client.HTTPRequest].
func (a *API) httpDelete(ctx context.Context, linkID, path string, query url.Values) error {
	_, err := a.httpRequest(ctx, linkID, path, http.MethodDelete, defaultAccept, query, nil)
	return err
}

// httpGet is a GitHub-specific HTTP GET wrapper for [client.HTTPRequest].
func (a *API) httpGet(ctx context.Context, linkID, path string, query url.Values, jsonResp any) (bool, error) {
	link, err := a.httpRequest(ctx, linkID, path, http.MethodGet, defaultAccept, query, jsonResp)
	return link != "", err
}

// httpPatch is a GitHub-specific HTTP PATCH wrapper for [client.HTTPRequest].
func (a *API) httpPatch(ctx context.Context, linkID, path, accept string, jsonBody, jsonResp any) error {
	_, err := a.httpRequest(ctx, linkID, path, http.MethodPatch, accept, jsonBody, jsonResp)
	return err
}

// httpPost is a GitHub-specific HTTP POST wrapper for [client.HTTPRequest].
func (a *API) httpPost(ctx context.Context, linkID, path, accept string, jsonBody, jsonResp any) error {
	_, err := a.httpRequest(ctx, linkID, path, http.MethodPost, accept, jsonBody, jsonResp)
	return err
}

// httpPut is a GitHub-specific HTTP PUT wrapper for [client.HTTPRequest].
func (a *API) httpPut(ctx context.Context, linkID, path, accept string, jsonBody, jsonResp any) error {
	_, err := a.httpRequest(ctx, linkID, path, http.MethodPut, accept, jsonBody, jsonResp)
	return err
}

func (a *API) httpRequest(ctx context.Context, linkID, path, method, accept string, queryOrJSONBody, parsedResp any) (string, error) {
	l, apiURL, auth, err := a.httpRequestPrep(ctx, linkID, path)
	if err != nil {
		return "", err
	}

	rawResp, headers, _, err := client.HTTPRequest(ctx, method, apiURL, auth, accept, client.ContentJSON, queryOrJSONBody)
	if err != nil {
		l.Error("HTTP request error", slog.Any("error", err), slog.String("http_method", method), slog.String("url", apiURL))
		return "", err
	}

	l.Info("sent HTTP request", slog.String("link_id", linkID), slog.String("http_method", method), slog.String("url", apiURL))

	if parsedResp == nil {
		return headers.Get("link"), nil // No response body expected.
	}

	if err := json.Unmarshal(rawResp, parsedResp); err != nil {
		msg := "failed to decode HTTP response's JSON body"
		l.Error(msg, slog.Any("error", err), slog.String("url", apiURL))
		msg = fmt.Sprintf("%s: %v", msg, err)
		return "", temporal.NewNonRetryableApplicationError(msg, fmt.Sprintf("%T", err), err, apiURL, string(rawResp))
	}

	return headers.Get("link"), nil
}

// httpRequestPrep supports custom Thrippy link IDs (for user impersonation).
// If it's empty, we use the Timpani server's preconfigured GitHub link ID.
func (a *API) httpRequestPrep(ctx context.Context, linkID, path string) (l log.Logger, apiURL, auth string, err error) {
	l = activity.GetLogger(ctx)

	var secrets map[string]string
	secrets, err = a.thrippy.LinkCreds(ctx, linkID)
	if err != nil {
		return l, "", "", err
	}

	baseURL := secrets["api_base_url"] // Added automatically to "github-app-jwt" links.
	if baseURL == "" {
		baseURL = secrets["base_url"] // Manual and optional in "github-app-user" and "github-user-pat" links.
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
		err = temporal.NewNonRetryableApplicationError(err.Error(), fmt.Sprintf("%T", err), err, baseURL, path)
		return l, "", "", err
	}

	// "access_token" has a value only in "github-app-user" link secrets.
	// "pat" has a value only in "github-user-pat" link secrets.
	if auth := secrets["access_token"] + secrets["pat"]; auth != "" {
		return l, apiURL, auth, nil
	}

	// Generating JWTs (and using them to generate installation tokens) is supported only for "github-app-jwt" links.
	auth, err = generateJWT(secrets["client_id"], secrets["private_key"])
	if err != nil {
		msg := "failed to generate JWT for GitHub API call"
		l.Warn(msg, slog.Any("error", err), slog.String("link_id", a.thrippy.LinkID))
		return l, "", "", temporal.NewNonRetryableApplicationError(msg, "error", err, a.thrippy.LinkID)
	}

	auth, err = a.createInstallationToken(ctx, baseURL, secrets["install_id"], auth)
	if err != nil {
		return l, "", "", err
	}

	return l, apiURL, auth, nil
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

// createInstallationToken retrieves a new installation access token for a GitHub app. Based on:
//   - https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/generating-an-installation-access-token-for-a-github-app
//   - https://docs.github.com/en/rest/apps/apps?apiVersion=2022-11-28#create-an-installation-access-token-for-an-app
func (a *API) createInstallationToken(ctx context.Context, baseURL, installID, auth string) (string, error) {
	l := activity.GetLogger(ctx)
	post := http.MethodPost

	tokenURL, err := url.JoinPath(baseURL, "/app/installations", installID, "access_tokens")
	if err != nil {
		l.Error("failed to construct GitHub installation access token URL", slog.Any("error", err),
			slog.String("base_url", baseURL), slog.String("install_id", installID))
		return "", err
	}

	rawResp, _, _, err := client.HTTPRequest(ctx, post, tokenURL, auth, defaultAccept, "", http.NoBody)
	if err != nil {
		l.Error("HTTP request error", slog.Any("error", err), slog.String("http_method", post), slog.String("url", tokenURL))
		return "", err
	}
	l.Info("sent HTTP request", slog.String("link_id", a.thrippy.LinkID),
		slog.String("http_method", post), slog.String("url", tokenURL))

	jsonResp := new(tokenResponse)
	if err := json.Unmarshal(rawResp, jsonResp); err != nil {
		l.Error("failed to decode GitHub installation access token response",
			slog.Any("error", err), slog.String("response", string(rawResp)))
		return "", err
	}

	return jsonResp.Token, nil
}
