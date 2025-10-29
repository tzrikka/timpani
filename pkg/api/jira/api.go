package jira

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
	URLPathPrefix = "/rest/api/3"
)

// httpGet is a Jira-specific HTTP GET wrapper for [client.HTTPRequest].
func (a *API) httpGet(ctx context.Context, pathSuffix string, query url.Values, jsonResp any) error {
	return a.httpRequest(ctx, pathSuffix, http.MethodGet, query, jsonResp)
}

func (a *API) httpRequest(ctx context.Context, pathSuffix, method string, queryOrJSONBody, jsonResp any) error {
	l, apiURL, auth, err := a.httpRequestPrep(ctx, pathSuffix)
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

func (a *API) httpRequestPrep(ctx context.Context, pathSuffix string) (l log.Logger, apiURL, auth string, err error) {
	l = activity.GetLogger(ctx)

	var secrets map[string]string
	secrets, err = a.thrippy.CustomLinkCreds(ctx, "")
	if err != nil {
		return
	}

	apiURL, err = url.JoinPath(secrets["base_url"], URLPathPrefix, pathSuffix)
	if err != nil {
		l.Error("failed to construct Jira API URL", "error", err, "base_url", secrets["base_url"], "path", URLPathPrefix+pathSuffix)
		err = temporal.NewNonRetryableApplicationError(err.Error(), fmt.Sprintf("%T", err), err)
		return
	}

	// "access_token" has a value only in "jira-app-oauth" link secrets.
	// The others have values only in "jira-user-token" link secrets.
	auth = secrets["access_token"]
	if auth == "" {
		auth = fmt.Sprintf("Basic %s:%s", secrets["email"], secrets["api_token"])
	}

	return
}
