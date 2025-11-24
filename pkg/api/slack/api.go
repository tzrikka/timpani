package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"

	"github.com/tzrikka/timpani/pkg/http/client"
)

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
		l.Error("failed to construct Slack API URL", "error", err, "base_url", baseURL, "url_suffix", urlSuffix)
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

	resp, retryAfter, err := client.HTTPRequest(ctx, http.MethodGet, apiURL, botToken, client.AcceptJSON, "", query)
	if err != nil {
		if retryAfter > 0 {
			l.Warn("throttling HTTP GET request", "retry_after", retryAfter, "url", apiURL)
			opts := temporal.ApplicationErrorOptions{NextRetryDelay: time.Second * time.Duration(retryAfter)}
			return temporal.NewApplicationErrorWithOptions(err.Error(), "RateLimitError", opts)
		}
		l.Error("HTTP GET request error", "error", err, "url", apiURL)
		return err
	}

	if err := json.Unmarshal(resp, jsonResp); err != nil {
		msg := "failed to decode HTTP response's JSON body"
		l.Error(msg, "error", err, "url", apiURL)
		msg = fmt.Sprintf("%s: %v", msg, err)
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

	resp, retryAfter, err := client.HTTPRequest(ctx, http.MethodPost, apiURL, botToken, client.AcceptJSON, client.ContentJSON, jsonBody)
	if err != nil {
		if retryAfter > 0 {
			l.Warn("throttling HTTP POST request", "retry_after", retryAfter, "url", apiURL)
			opts := temporal.ApplicationErrorOptions{NextRetryDelay: time.Second * time.Duration(retryAfter)}
			return temporal.NewApplicationErrorWithOptions(err.Error(), "RateLimitError", opts)
		}
		l.Error("HTTP POST request error", "error", err, "url", apiURL)
		return err
	}

	if err := json.Unmarshal(resp, jsonResp); err != nil {
		msg := "failed to decode HTTP response's JSON body"
		l.Error(msg, "error", err, "url", apiURL)
		msg = fmt.Sprintf("%s: %v", msg, err)
		return temporal.NewNonRetryableApplicationError(msg, fmt.Sprintf("%T", err), err, apiURL)
	}

	l.Info("sent HTTP POST request", "link_id", a.thrippy.LinkID, "url", apiURL)
	return nil
}

// httpPostFile is an HTTP POST wrapper of [client.HTTPRequest] for uploading files to Slack.
func (a *API) httpPostFile(ctx context.Context, uploadURL, contentType string, content []byte) error {
	l := activity.GetLogger(ctx)

	if resp, _, err := client.HTTPRequest(ctx, http.MethodPost, uploadURL, "", "", contentType, content); err != nil {
		l.Error("HTTP POST request error", "error", err, "url", uploadURL, "content_type", contentType, "response", string(resp))
		return err
	}

	l.Info("sent HTTP POST request", "url", uploadURL, "content_type", contentType, "length", len(content))
	return nil
}

// slackAPIError serializes a Slack API error response into an error.
func slackAPIError(resp any, errCode string) error {
	sb := new(strings.Builder)
	if err := json.NewEncoder(sb).Encode(resp); err != nil {
		return errors.New(errCode)
	}
	return errors.New(strings.TrimSpace(sb.String()))
}
