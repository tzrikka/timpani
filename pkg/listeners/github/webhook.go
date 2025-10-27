// Package github implements an HTTP webhook to handle
// GitHub events (https://docs.github.com/en/webhooks).
package github

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/tzrikka/timpani/internal/listeners"
	"github.com/tzrikka/timpani/pkg/metrics"
	"github.com/tzrikka/timpani/pkg/temporal"
)

const (
	contentTypeHeader = "Content-Type"
	eventHeader       = "X-Github-Event"
	signatureHeader   = "X-Hub-Signature-256"
)

func WebhookHandler(ctx context.Context, _ http.ResponseWriter, r listeners.RequestData) int {
	l := zerolog.Ctx(ctx).With().Str("link_type", "github").Str("link_medium", "webhook").Logger()
	t := time.Now().UTC()

	if statusCode := checkContentTypeHeader(l, r); statusCode != http.StatusOK {
		return metrics.CountWebhookEvent(l, t, "", statusCode)
	}
	if statusCode := CheckSignatureHeader(l, r); statusCode != http.StatusOK {
		return metrics.CountWebhookEvent(l, t, "", statusCode)
	}

	// If the payload is a web form, convert it to JSON.
	if r.Headers.Get(contentTypeHeader) == "application/x-www-form-urlencoded" {
		reader := strings.NewReader(r.WebForm.Get("payload"))
		if err := json.NewDecoder(reader).Decode(&r.JSONPayload); err != nil {
			l.Err(err).Msg("failed to extract and decode JSON payload from form data")
			return metrics.CountWebhookEvent(l, t, "", http.StatusInternalServerError)
		}
	}

	// Dispatch the event notification as a Temporal signal.
	signalName := "github.events." + r.Headers.Get(eventHeader)
	if err := temporal.Signal(ctx, r.Temporal, signalName, r.JSONPayload); err != nil {
		l.Err(err).Msg("failed to send Temporal signal")
		return metrics.CountWebhookEvent(l, t, signalName, http.StatusInternalServerError)
	}

	return metrics.CountWebhookEvent(l, t, signalName, http.StatusOK)
}

func checkContentTypeHeader(l zerolog.Logger, r listeners.RequestData) int {
	expected := []string{"application/json", "application/x-www-form-urlencoded"}
	ct := r.Headers.Get(contentTypeHeader)

	if !slices.Contains(expected, ct) {
		l.Warn().Str("header", contentTypeHeader).Str("got", ct).Any("want", expected).
			Msg("bad request: unexpected header value")
		return http.StatusBadRequest
	}

	return http.StatusOK
}

// CheckSignatureHeader is defined by and for GitHub, but also reused by Bitbucket.
func CheckSignatureHeader(l zerolog.Logger, r listeners.RequestData) int {
	sig := r.Headers.Get(signatureHeader)
	if sig == "" {
		l.Warn().Str("header", signatureHeader).Msg("bad request: missing header")
		return http.StatusForbidden
	}

	secret := r.LinkSecrets["webhook_secret"]
	if secret == "" {
		l.Warn().Msg("webhook secret is not configured")
		return http.StatusInternalServerError
	}

	if !verifySignature(l, secret, sig, r.RawPayload) {
		l.Warn().Str("signature", sig).Bool("has_signing_secret", secret != "").
			Msg("signature verification failed")
		return http.StatusForbidden
	}

	return http.StatusOK
}

// verifySignature implements
// https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries.
func verifySignature(l zerolog.Logger, webhookSecret, want string, body []byte) bool {
	mac := hmac.New(sha256.New, []byte(webhookSecret))

	n, err := mac.Write(body)
	if err != nil {
		l.Err(err).Msg("HMAC write error")
		return false
	}
	if n != len(body) {
		return false
	}

	got := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(got), []byte(want))
}
