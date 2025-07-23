package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"

	"github.com/tzrikka/timpani/pkg/http/client"
)

type slackResponse struct {
	OK               bool              `json:"ok"`
	Error            string            `json:"error,omitempty"`
	Needed           string            `json:"needed,omitempty"`   // Scope errors (undocumented).
	Provided         string            `json:"provided,omitempty"` // Scope errors (undocumented).
	Warning          string            `json:"warning,omitempty"`
	ResponseMetadata *responseMetadata `json:"response_metadata,omitempty"`
}

type responseMetadata struct {
	Messages   []string `json:"messages,omitempty"`
	Warnings   []string `json:"warnings,omitempty"`
	NextCursor string   `json:"next_cursor,omitempty"`
}

func (a *API) httpRequestPrep(ctx context.Context, urlSuffix string) (l log.Logger, apiURL, botToken string, err error) {
	l = activity.GetLogger(ctx)

	var template string
	var secrets map[string]string
	template, secrets, err = a.thrippy.LinkData(ctx, "slack")
	if err != nil {
		return
	}

	urlBase := "https://slack.com"
	if template == "slack-oauth-gov" {
		urlBase = "https://slack-gov.com" // https://docs.slack.dev/govslack
	}

	apiURL, err = url.JoinPath(urlBase, "api", strings.TrimPrefix(urlSuffix, "slack."))
	if err != nil {
		msg := "failed to construct Slack API URL"
		l.Error(msg, "error", err.Error(), "url_base", urlBase, "url_suffix", urlSuffix)
		err = temporal.NewNonRetryableApplicationError(err.Error(), fmt.Sprintf("%T", err), err)
		return
	}

	botToken = secrets["bot_token"]
	if botToken == "" {
		botToken = secrets["access_token"] // OAuth token, possibly short-lived.
	}
	if botToken == "" {
		msg := "Slack bot token not found in Thrippy link credentials"
		l.Warn(msg, "link_id", a.thrippy.LinkID)
		err = temporal.NewNonRetryableApplicationError(msg, "error", nil, a.thrippy.LinkID)
		return
	}

	return
}

func (a *API) httpGet(ctx context.Context, urlSuffix string, query url.Values, jsonResp any) error {
	l, apiURL, botToken, err := a.httpRequestPrep(ctx, urlSuffix)
	if err != nil {
		return err
	}

	resp, err := client.HTTPRequest(ctx, http.MethodGet, apiURL, botToken, query)
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

	l.Info("successful HTTP GET request", "link_id", a.thrippy.LinkID, "url", apiURL)
	return nil
}

func (a *API) httpPost(ctx context.Context, urlSuffix string, jsonBody, jsonResp any) error {
	l, apiURL, botToken, err := a.httpRequestPrep(ctx, urlSuffix)
	if err != nil {
		return err
	}

	resp, err := client.HTTPRequest(ctx, http.MethodPost, apiURL, botToken, jsonBody)
	if err != nil {
		l.Error("HTTP POST request error", "error", err.Error(), "url", apiURL)
		return err
	}

	if err := json.Unmarshal(resp, jsonResp); err != nil {
		msg := "failed to decode HTTP response's JSON body"
		l.Error(msg, "error", err.Error(), "url", apiURL)
		msg = fmt.Sprintf("%s: %s", msg, err.Error())
		return temporal.NewNonRetryableApplicationError(msg, fmt.Sprintf("%T", err), err, apiURL)
	}

	l.Info("successful HTTP POST request", "link_id", a.thrippy.LinkID, "url", apiURL)
	return nil
}
