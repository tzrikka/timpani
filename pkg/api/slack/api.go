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
	template, secrets, err = a.thrippy.LinkData(ctx)
	if err != nil {
		return
	}

	baseURL := "https://slack.com"
	if template == "slack-oauth-gov" {
		baseURL = "https://slack-gov.com" // https://docs.slack.dev/govslack
	}

	apiURL, err = url.JoinPath(baseURL, "api", strings.TrimPrefix(urlSuffix, "slack."))
	if err != nil {
		msg := "failed to construct Slack API URL"
		l.Error(msg, "error", err.Error(), "base_url", baseURL, "url_suffix", urlSuffix)
		err = temporal.NewNonRetryableApplicationError(err.Error(), fmt.Sprintf("%T", err), err)
		return
	}

	botToken = secrets["bot_token"]
	if botToken == "" {
		botToken = secrets["access_token"] // Short-lived OAuth token.
	}
	if botToken == "" {
		msg := "Slack bot token not found in Thrippy link credentials"
		l.Warn(msg, "link_id", a.thrippy.LinkID)
		err = temporal.NewNonRetryableApplicationError(msg, "error", nil, a.thrippy.LinkID)
		return
	}

	return
}

// httpGet is a Slack-specific HTTP GET wrapper for [client.HTTPRequest].
func (a *API) httpGet(ctx context.Context, urlSuffix string, query url.Values, jsonResp any) error {
	l, apiURL, botToken, err := a.httpRequestPrep(ctx, urlSuffix)
	if err != nil {
		return err
	}

	resp, err := client.HTTPRequest(ctx, http.MethodGet, apiURL, botToken, client.AcceptJSON, query)
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

// httpPost is a Slack-specific HTTP POST wrapper for [client.HTTPRequest].
func (a *API) httpPost(ctx context.Context, urlSuffix string, jsonBody, jsonResp any) error {
	l, apiURL, botToken, err := a.httpRequestPrep(ctx, urlSuffix)
	if err != nil {
		return err
	}

	resp, err := client.HTTPRequest(ctx, http.MethodPost, apiURL, botToken, client.AcceptJSON, jsonBody)
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

	l.Info("sent HTTP POST request", "link_id", a.thrippy.LinkID, "url", apiURL)
	return nil
}
