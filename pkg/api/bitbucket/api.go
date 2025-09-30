package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"

	"github.com/tzrikka/timpani/pkg/http/client"
)

const (
	BaseURL = "https://api.bitbucket.org/2.0"
)

// httpDelete is a Bitbucket-specific HTTP DELETE wrapper for [client.HTTPRequest].
func (a *API) httpDelete(ctx context.Context, linkID, path string, query url.Values) error {
	return a.httpRequest(ctx, linkID, path, http.MethodDelete, query, nil)
}

// httpGet is a Bitbucket-specific HTTP GET wrapper for [client.HTTPRequest].
func (a *API) httpGet(ctx context.Context, linkID, path string, query url.Values, jsonResp any) error {
	return a.httpRequest(ctx, linkID, path, http.MethodGet, query, jsonResp)
}

// httpPost is a Bitbucket-specific HTTP POST wrapper for [client.HTTPRequest].
func (a *API) httpPost(ctx context.Context, linkID, path string, jsonBody, jsonResp any) error {
	return a.httpRequest(ctx, linkID, path, http.MethodPost, jsonBody, jsonResp)
}

// httpPut is a Bitbucket-specific HTTP PUT wrapper for [client.HTTPRequest].
func (a *API) httpPut(ctx context.Context, linkID, path string, jsonBody, jsonResp any) error {
	return a.httpRequest(ctx, linkID, path, http.MethodPut, jsonBody, jsonResp)
}

func (a *API) httpRequest(ctx context.Context, linkID, path, method string, queryOrJSONBody, jsonResp any) error {
	l, apiURL, auth, err := a.httpRequestPrep(ctx, linkID, path)
	if err != nil {
		return err
	}

	resp, err := client.HTTPRequest(ctx, method, apiURL, auth, client.AcceptJSON, queryOrJSONBody)
	if err != nil {
		l.Error("HTTP request error", "method", method, "error", err, "url", apiURL)
		return err
	}

	l.Info("sent HTTP request", "link_id", a.thrippy.LinkID, "method", method, "url", apiURL)
	if jsonResp == nil {
		return nil
	}

	if err := json.Unmarshal(resp, jsonResp); err != nil {
		msg := "failed to decode HTTP response's JSON body"
		l.Error(msg, "error", err, "url", apiURL)
		msg = fmt.Sprintf("%s: %v", msg, err)
		return temporal.NewNonRetryableApplicationError(msg, fmt.Sprintf("%T", err), err, apiURL)
	}

	return nil
}

// httpRequestPrep supports custom Thrippy link IDs (for user impersonation).
// If it's empty, we use the Timpani server's preconfigured Bitbucket link ID.
func (a *API) httpRequestPrep(ctx context.Context, linkID, path string) (l log.Logger, apiURL, auth string, err error) {
	l = activity.GetLogger(ctx)

	var secrets map[string]string
	secrets, err = a.thrippy.CustomLinkCreds(ctx, linkID)
	if err != nil {
		return
	}

	apiURL, err = url.JoinPath(BaseURL, path)
	if err != nil {
		l.Error("failed to construct Bitbucket API URL", "error", err, "base_url", BaseURL, "path", path)
		err = temporal.NewNonRetryableApplicationError(err.Error(), fmt.Sprintf("%T", err), err)
		return
	}

	// "access_token" has a value only in "bitbucket-app-oauth" link secrets.
	// The others have values only in "bitbucket-user-token" link secrets.
	auth = secrets["access_token"]
	if auth == "" {
		auth = fmt.Sprintf("Basic %s:%s", secrets["email"], secrets["api_token"])
	}

	return
}
