package slack

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/tzrikka/timpani/internal/listeners"
	"github.com/tzrikka/timpani/internal/logger"
	"github.com/tzrikka/timpani/pkg/http/client"
	"github.com/tzrikka/timpani/pkg/otel"
)

const (
	contentTypeHeader = "Content-Type"
	timestampHeader   = "X-Slack-Request-Timestamp"
	signatureHeader   = "X-Slack-Signature"

	// The maximum shift/delay that we allow between an inbound request's
	// timestamp, and our current timestamp, to defend against replay attacks.
	// See https://docs.slack.dev/authentication/verifying-requests-from-slack.
	maxDifference = 5 * time.Minute

	// Slack API implementation detail.
	// See https://docs.slack.dev/authentication/verifying-requests-from-slack.
	slackSigVersion = "v0"
)

type slashCommandResponse struct {
	ResponseType string `json:"response_type"`
	Text         string `json:"text"`
}

func WebhookHandler(ctx context.Context, w http.ResponseWriter, r listeners.RequestData) int {
	l := logger.FromContext(ctx).With(slog.String("link_type", "slack"), slog.String("link_medium", "webhook"))
	t := time.Now().UTC()

	if statusCode := checkContentTypeHeader(l, r); statusCode != http.StatusOK {
		return otel.IncrementWebhookEventCounter(l, t, "", statusCode)
	}
	if statusCode := checkTimestampHeader(l, r); statusCode != http.StatusOK {
		return otel.IncrementWebhookEventCounter(l, t, "", statusCode)
	}
	if statusCode := checkSignatureHeader(l, r); statusCode != http.StatusOK {
		return otel.IncrementWebhookEventCounter(l, t, "", statusCode)
	}

	// Special handling for some events.

	// https://docs.slack.dev/reference/events/url_verification
	if r.JSONPayload["type"] == "url_verification" {
		l.Debug("replied to Slack URL verification event", slog.String("event_type", "url_verification"))
		w.Header().Add(contentTypeHeader, "text/plain")
		_, _ = fmt.Fprint(w, r.JSONPayload["challenge"])

		otel.IncrementWebhookEventCounter(l, t, "slack.events.url_verification", http.StatusOK)
		return 0 // [http.StatusOK] already written by "w.Write" ("fmt.Fprint(w)").
	}

	// https://docs.slack.dev/interactivity/implementing-slash-commands#command_payload_descriptions
	// (the informational note under the payload info table).
	if r.WebForm.Get("ssl_check") != "" {
		return otel.IncrementWebhookEventCounter(l, t, "slack.events.ssl_check", http.StatusOK)
	}

	// https://docs.slack.dev/interactivity/implementing-slash-commands#responding_to_commands
	// https://docs.slack.dev/interactivity/implementing-slash-commands#responding_with_errors
	// https://docs.slack.dev/interactivity/implementing-slash-commands#best-practices
	statusCode := http.StatusOK
	if sc := r.WebForm.Get("command"); sc != "" {
		l.Debug("replied to Slack slash command", slog.String("event_type", "slash_command"))
		w.Header().Add(contentTypeHeader, "application/json; charset=utf-8")

		text := fmt.Sprintf("Your command: `%s %s`", sc, r.WebForm.Get("text"))
		resp := slashCommandResponse{ResponseType: "ephemeral", Text: text}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			l.Error("failed to encode JSON response", slog.Any("error", err))
		}

		statusCode = 0 // [http.StatusOK] already written by "w.Write".
	}

	// Dispatch the event notification, based on its type.
	signalName, err := dispatchFromWebhook(logger.WithContext(ctx, l), r)
	if err != nil {
		return otel.IncrementWebhookEventCounter(l, t, signalName, http.StatusInternalServerError)
	}

	return otel.IncrementWebhookEventCounter(l, t, signalName, statusCode)
}

func checkContentTypeHeader(l *slog.Logger, r listeners.RequestData) int {
	expected := []string{"application/json", client.ContentForm}
	ct := r.Headers.Get(contentTypeHeader)

	if !slices.Contains(expected, ct) {
		l.Warn("bad request: unexpected header value", slog.String("header", contentTypeHeader),
			slog.String("got", ct), slog.Any("want", expected))
		return http.StatusBadRequest
	}

	return http.StatusOK
}

func checkTimestampHeader(l *slog.Logger, r listeners.RequestData) int {
	ts := r.Headers.Get(timestampHeader)
	if ts == "" {
		l.Warn("bad request: missing header", slog.String("header", timestampHeader))
		return http.StatusBadRequest
	}

	secs, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		l.Warn("bad request: invalid header value", slog.String("header", timestampHeader),
			slog.String("got", ts))
		return http.StatusBadRequest
	}

	d := time.Since(time.Unix(secs, 0))
	if d.Abs() > maxDifference {
		l.Warn("bad request: stale header value", slog.String("header", timestampHeader),
			slog.Duration("difference", d))
		return http.StatusBadRequest
	}

	return http.StatusOK
}

func checkSignatureHeader(l *slog.Logger, r listeners.RequestData) int {
	sig := r.Headers.Get(signatureHeader)
	if sig == "" {
		l.Warn("bad request: missing header", slog.String("header", signatureHeader))
		return http.StatusForbidden
	}

	secret := r.LinkSecrets["signing_secret"]
	if secret == "" {
		l.Warn("signing secret is not configured")
		return http.StatusInternalServerError
	}

	ts := r.Headers.Get(timestampHeader)
	if !verifySignature(l, secret, ts, sig, r.RawPayload) {
		l.Warn("signature verification failed", slog.String("signature", sig),
			slog.Bool("has_signing_secret", secret != ""))
		return http.StatusForbidden
	}

	return http.StatusOK
}

// verifySignature implements
// https://docs.slack.dev/authentication/verifying-requests-from-slack.
func verifySignature(l *slog.Logger, signingSecret, ts, want string, body []byte) bool {
	mac := hmac.New(sha256.New, []byte(signingSecret))

	n, err := mac.Write(fmt.Appendf(nil, "%s:%s:", slackSigVersion, ts))
	if err != nil {
		l.Error("HMAC write error", slog.Any("error", err))
		return false
	}
	if n != len(ts)+4 {
		return false
	}

	if n, err := mac.Write(body); err != nil || n != len(body) {
		return false
	}

	got := fmt.Sprintf("%s=%s", slackSigVersion, hex.EncodeToString(mac.Sum(nil)))
	return hmac.Equal([]byte(got), []byte(want))
}
