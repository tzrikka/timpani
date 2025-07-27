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

	"github.com/rs/zerolog"

	"github.com/tzrikka/timpani/internal/listeners"
	"github.com/tzrikka/timpani/pkg/temporal"
)

const (
	contentTypeHeader = "Content-Type"
	eventHeader       = "X-Github-Event"
	signatureHeader   = "X-Hub-Signature-256"
)

func WebhookHandler(ctx context.Context, _ http.ResponseWriter, r listeners.RequestData) int {
	l := zerolog.Ctx(ctx).With().Str("link_type", "github").Str("link_medium", "webhook").Logger()

	statusCode := checkContentTypeHeader(l, r)
	if statusCode != http.StatusOK {
		return statusCode
	}

	statusCode = checkSignatureHeader(l, r)
	if statusCode != http.StatusOK {
		return statusCode
	}

	// If the payload is a web form, convert it to JSON.
	if r.Headers.Get(contentTypeHeader) == "application/x-www-form-urlencoded" {
		reader := strings.NewReader(r.WebForm.Get("payload"))
		if err := json.NewDecoder(reader).Decode(&r.JSONPayload); err != nil {
			l.Err(err).Msg("failed to extract and decode JSON payload from form data")
			return http.StatusInternalServerError
		}
	}

	// Dispatch the event notification as a Temporal signal.
	signalName := "github.events." + r.Headers.Get(eventHeader)
	if err := temporal.Signal(ctx, r.Temporal, signalName, r.JSONPayload); err != nil {
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

func checkSignatureHeader(l zerolog.Logger, r listeners.RequestData) int {
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
