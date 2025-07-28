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
	baseURL = "https://api.bitbucket.org/2.0"
)

func (a *API) httpRequestPrep(ctx context.Context, path string) (l log.Logger, apiURL, auth string, err error) {
	l = activity.GetLogger(ctx)

	var secrets map[string]string
	secrets, err = a.thrippy.LinkCreds(ctx)
	if err != nil {
		return
	}

	apiURL, err = url.JoinPath(baseURL, path)
	if err != nil {
		msg := "failed to construct GitHub API URL"
		l.Error(msg, "error", err.Error(), "base_url", baseURL, "path", path)
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

// httpGet is a Bitbucket-specific HTTP GET wrapper for [client.HTTPRequest].
func (a *API) httpGet(ctx context.Context, path string, query url.Values, jsonResp any) error {
	l, apiURL, auth, err := a.httpRequestPrep(ctx, path)
	if err != nil {
		return err
	}

	resp, err := client.HTTPRequest(ctx, http.MethodGet, apiURL, auth, client.AcceptJSON, query)
	if err != nil {
		l.Error("HTTP GET request error", "error", err.Error(), "url", apiURL)
		return err
	}

	if err := json.Unmarshal(resp, jsonResp); err != nil {
		msg := "failed to decode HTTP response's JSON body"
		l.Error(msg, "error", err.Error(), "url", apiURL)
		msg = fmt.Sprintf("%s: %s", msg, err.Error())
		return temporal.NewNonRetryableApplicationError(msg, fmt.Sprintf("%T", err), err, apiURL)
	}

	l.Info("sent HTTP GET request", "link_id", a.thrippy.LinkID, "url", apiURL)
	return nil
}
