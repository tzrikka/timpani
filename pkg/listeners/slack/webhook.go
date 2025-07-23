package slack

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/tzrikka/timpani/internal/listeners"
	"github.com/tzrikka/timpani/pkg/temporal"
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

func WebhookHandler(ctx context.Context, w http.ResponseWriter, r listeners.RequestData) int {
	l := zerolog.Ctx(ctx).With().Str("link_type", "slack").Str("link_medium", "webhook").Logger()

	statusCode := checkContentTypeHeader(l, r)
	if statusCode != http.StatusOK {
		return statusCode
	}

	statusCode = checkTimestampHeader(l, r)
	if statusCode != http.StatusOK {
		return statusCode
	}

	statusCode = checkSignatureHeader(l, r)
	if statusCode != http.StatusOK {
		return statusCode
	}

	// Special handling for some events.

	// https://docs.slack.dev/reference/events/url_verification
	if r.JSONPayload["type"] == "url_verification" {
		l.Debug().Str("event_type", "url_verification").
			Msg("replied to Slack URL verification event")
		w.Header().Add(contentTypeHeader, "text/plain")
		_, _ = fmt.Fprint(w, r.JSONPayload["challenge"])
		return 0 // [http.StatusOK] already written by "w.Write" ("fmt.Fprint(w)").
	}

	// https://docs.slack.dev/interactivity/implementing-slash-commands#command_payload_descriptions
	// (the informational note under the payload info table).
	if r.WebForm.Get("ssl_check") != "" {
		return http.StatusOK
	}

	// https://docs.slack.dev/interactivity/implementing-slash-commands#responding_to_commands
	// https://docs.slack.dev/interactivity/implementing-slash-commands#responding_with_errors
	// https://docs.slack.dev/interactivity/implementing-slash-commands#best-practices
	if sc := r.WebForm.Get("command"); sc != "" {
		l.Debug().Str("event_type", "slash_command").Msg("replied to Slack slash command")
		w.Header().Add(contentTypeHeader, "application/json; charset=utf-8")
		resp := "{\"response_type\": \"ephemeral\", \"text\": \"Your command: `%s %s`\"}"
		_, _ = fmt.Fprintf(w, resp, sc, r.WebForm.Get("text"))
		return 0 // [http.StatusOK] already written by "w.Write" ("fmt.Fprintf(w)").
	}

	// Dispatch the event notification, based on its type.

	// https://docs.slack.dev/apis/events-api#events-JSON
	payload := r.JSONPayload
	eventType := payload["type"]
	if eventType == "event_callback" {
		if m, ok := payload["event"].(map[string]any); ok {
			eventType = m["type"]
		}
	}

	// https://docs.slack.dev/interactivity/implementing-slash-commands#app_command_handling
	if r.WebForm.Get("command") != "" {
		eventType = "slash_command"
		payload = webFormToMap(r.WebForm)
	}

	// https://docs.slack.dev/interactivity/handling-user-interaction#payloads
	// https://docs.slack.dev/reference/interaction-payloads
	if p := r.WebForm.Get("payload"); p != "" {
		payload = map[string]any{}
		if err := json.NewDecoder(strings.NewReader(p)).Decode(&payload); err != nil {
			l.Err(err).Msg("failed to decode interaction event payload")
			return http.StatusInternalServerError
		}
		eventType = payload["type"]
	}

	ctx = l.WithContext(ctx)
	name := fmt.Sprintf("slack.events.%s", eventType)
	if err := temporal.Signal(ctx, r.Temporal, name, payload); err != nil {
		l.Err(err).Msg("failed to send Temporal signal")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func checkContentTypeHeader(l zerolog.Logger, r listeners.RequestData) int {
	expected := []string{"application/json", "application/x-www-form-urlencoded"}
	v := r.Headers.Get(contentTypeHeader)

	if !slices.Contains(expected, v) {
		l.Warn().Str("header", contentTypeHeader).Str("got", v).Any("want", expected).
			Msg("bad request: unexpected header value")
		return http.StatusBadRequest
	}

	return http.StatusOK
}

func checkTimestampHeader(l zerolog.Logger, r listeners.RequestData) int {
	ts := r.Headers.Get(timestampHeader)
	if ts == "" {
		l.Warn().Str("header", timestampHeader).Msg("bad request: missing header")
		return http.StatusBadRequest
	}

	secs, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		l.Warn().Str("header", timestampHeader).Str("got", ts).
			Msg("bad request: invalid header value")
		return http.StatusBadRequest
	}

	d := time.Since(time.Unix(secs, 0))
	if d.Abs() > maxDifference {
		l.Warn().Str("header", timestampHeader).Dur("difference", d).
			Msg("bad request: stale header value")
		return http.StatusBadRequest
	}

	return http.StatusOK
}

func checkSignatureHeader(l zerolog.Logger, r listeners.RequestData) int {
	sig := r.Headers.Get(signatureHeader)
	if sig == "" {
		l.Warn().Str("header", signatureHeader).Msg("bad request: missing header")
		return http.StatusForbidden
	}

	secret := r.LinkSecrets["signing_secret"]
	if secret == "" {
		l.Warn().Msg("signing secret is not configured")
		return http.StatusInternalServerError
	}

	ts := r.Headers.Get(timestampHeader)
	if !verifySignature(l, secret, ts, sig, r.RawPayload) {
		l.Warn().Str("signature", sig).Bool("has_signing_secret", secret != "").
			Msg("signature verification failed")
		return http.StatusForbidden
	}

	return http.StatusOK
}

// verifySignature implements
// https://docs.slack.dev/authentication/verifying-requests-from-slack.
func verifySignature(l zerolog.Logger, signingSecret, ts, want string, body []byte) bool {
	mac := hmac.New(sha256.New, []byte(signingSecret))

	n, err := mac.Write(fmt.Appendf(nil, "%s:%s:", slackSigVersion, ts))
	if err != nil {
		l.Err(err).Msg("HMAC write error")
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

// webFormToMap converts a web form into a Go map which is compatible with JSON.
func webFormToMap(vs url.Values) map[string]any {
	m := make(map[string]any)
	for k, v := range vs {
		m[k] = v[0]
	}
	return m
}
