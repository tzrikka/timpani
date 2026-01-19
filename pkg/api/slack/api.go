package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"

	"github.com/tzrikka/timpani-api/pkg/slack"
	"github.com/tzrikka/timpani/pkg/http/client"
	"github.com/tzrikka/timpani/pkg/metrics"
)

func (a *API) httpRequestPrep(ctx context.Context, urlSuffix string) (l log.Logger, t time.Time, apiURL, botToken string, err error) {
	l = activity.GetLogger(ctx)
	t = time.Now().UTC()

	var template string
	var secrets map[string]string
	template, secrets, err = a.thrippy.LinkData(ctx)
	if err != nil {
		return l, t, "", "", err
	}

	baseURL := "https://slack.com"
	if template == "slack-oauth-gov" {
		baseURL = "https://slack-gov.com" // https://docs.slack.dev/govslack
	}

	suffix := strings.TrimPrefix(urlSuffix, "slack.")
	apiURL, err = url.JoinPath(baseURL, "api", suffix)
	if err != nil {
		l.Error("failed to construct Slack API URL", slog.Any("error", err),
			slog.String("base_url", baseURL), slog.String("url_suffix", urlSuffix))
		err = temporal.NewNonRetryableApplicationError(err.Error(), fmt.Sprintf("%T", err), err, baseURL, "api", suffix)
		return l, t, "", "", err
	}

	botToken = secrets["bot_token"]
	if botToken == "" {
		botToken = secrets["access_token"] // Short-lived OAuth token.
	}
	if botToken == "" {
		msg := "Slack bot token not found in Thrippy link credentials"
		l.Warn(msg, slog.String("link_id", a.thrippy.LinkID))
		err = temporal.NewNonRetryableApplicationError(msg, "error", nil, a.thrippy.LinkID)
		return l, t, apiURL, botToken, err
	}

	return l, t, apiURL, botToken, nil
}

// httpGet is a Slack-specific HTTP GET wrapper for [client.HTTPRequest].
func (a *API) httpGet(ctx context.Context, urlSuffix string, query url.Values, jsonResp any) error {
	l, t, apiURL, botToken, err := a.httpRequestPrep(ctx, urlSuffix)
	if err != nil {
		return err
	}

	resp, _, retryAfter, err := client.HTTPRequest(ctx, http.MethodGet, apiURL, botToken, client.AcceptJSON, "", query)
	if err != nil {
		metrics.IncrementAPICallCounter(t, urlSuffix, err)

		if retryAfter > 0 {
			l.Warn("throttling HTTP GET request", slog.Int("retry_after", retryAfter), slog.String("url", apiURL))
			opts := temporal.ApplicationErrorOptions{NextRetryDelay: time.Second * time.Duration(retryAfter)}
			return temporal.NewApplicationErrorWithOptions(err.Error(), "RateLimitError", opts)
		}

		l.Error("HTTP GET request error", slog.Any("error", err), slog.String("url", apiURL))
		return err
	}

	if err := json.Unmarshal(resp, jsonResp); err != nil {
		metrics.IncrementAPICallCounter(t, urlSuffix, err)

		msg := "failed to decode HTTP GET response's JSON body"
		l.Error(msg, slog.Any("error", err), slog.String("url", apiURL))
		msg = fmt.Sprintf("%s: %v", msg, err)
		return temporal.NewNonRetryableApplicationError(msg, fmt.Sprintf("%T", err), err, apiURL, string(resp))
	}

	baseResp := new(slack.Response)
	if err := json.Unmarshal(resp, baseResp); err == nil && !baseResp.OK && strings.Contains(baseResp.Error, "invalid") {
		err = errors.New(string(resp))
		metrics.IncrementAPICallCounter(t, urlSuffix, err)
		return temporal.NewNonRetryableApplicationError(baseResp.Error, "SlackAPIError", err, jsonResp)
	}

	l.Info("sent HTTP GET request", slog.String("link_id", a.thrippy.LinkID), slog.String("url", apiURL))
	metrics.IncrementAPICallCounter(t, urlSuffix, err)
	return nil
}

// httpPost is a Slack-specific HTTP POST wrapper for [client.HTTPRequest].
func (a *API) httpPost(ctx context.Context, urlSuffix string, jsonBody, jsonResp any) error {
	l, t, apiURL, botToken, err := a.httpRequestPrep(ctx, urlSuffix)
	if err != nil {
		return err
	}

	resp, _, retryAfter, err := client.HTTPRequest(ctx, http.MethodPost, apiURL, botToken, client.AcceptJSON, client.ContentJSON, jsonBody)
	if err != nil {
		metrics.IncrementAPICallCounter(t, urlSuffix, err)

		if retryAfter > 0 {
			l.Warn("throttling HTTP POST request", slog.Int("retry_after", retryAfter), slog.String("url", apiURL))
			opts := temporal.ApplicationErrorOptions{NextRetryDelay: time.Second * time.Duration(retryAfter)}
			return temporal.NewApplicationErrorWithOptions(err.Error(), "RateLimitError", opts)
		}

		l.Error("HTTP POST request error", slog.Any("error", err), slog.String("url", apiURL))
		return err
	}

	if err := json.Unmarshal(resp, jsonResp); err != nil {
		metrics.IncrementAPICallCounter(t, urlSuffix, err)

		msg := "failed to decode HTTP POST response's JSON body"
		l.Error(msg, slog.Any("error", err), slog.String("url", apiURL))
		msg = fmt.Sprintf("%s: %v", msg, err)
		return temporal.NewNonRetryableApplicationError(msg, fmt.Sprintf("%T", err), err, apiURL, string(resp))
	}

	baseResp := new(slack.Response)
	if err := json.Unmarshal(resp, baseResp); err == nil && !baseResp.OK && strings.Contains(baseResp.Error, "invalid") {
		err = errors.New(string(resp))
		metrics.IncrementAPICallCounter(t, urlSuffix, err)
		return temporal.NewNonRetryableApplicationError(baseResp.Error, "SlackAPIError", err, jsonResp)
	}

	l.Info("sent HTTP POST request", slog.String("link_id", a.thrippy.LinkID), slog.String("url", apiURL))
	metrics.IncrementAPICallCounter(t, urlSuffix, err)
	return nil
}

// httpPostFile is an HTTP POST wrapper of [client.HTTPRequest] for uploading files to Slack.
func (a *API) httpPostFile(ctx context.Context, uploadURL, contentType string, content []byte) error {
	l := activity.GetLogger(ctx)
	t := time.Now().UTC()

	if resp, _, _, err := client.HTTPRequest(ctx, http.MethodPost, uploadURL, "", "", contentType, content); err != nil {
		l.Error("HTTP POST request error", slog.Any("error", err), slog.String("url", uploadURL),
			slog.String("content_type", contentType), slog.String("response", string(resp)))
		metrics.IncrementAPICallCounter(t, slack.TimpaniUploadExternalActivityName, err)
		return err
	}

	l.Info("sent HTTP POST request", slog.String("url", uploadURL),
		slog.String("content_type", contentType), slog.Int("length", len(content)))
	metrics.IncrementAPICallCounter(t, slack.TimpaniUploadExternalActivityName, nil)
	return nil
}
