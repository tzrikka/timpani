// Package github implements an HTTP webhook to handle
// GitHub events (https://docs.github.com/en/webhooks).
package github

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/tzrikka/timpani/internal/listeners"
	"github.com/tzrikka/timpani/internal/logger"
	"github.com/tzrikka/timpani/pkg/http/client"
	"github.com/tzrikka/timpani/pkg/otel"
	"github.com/tzrikka/timpani/pkg/temporal"
)

const (
	contentTypeHeader = "Content-Type"
	eventHeader       = "X-Github-Event"
	signatureHeader   = "X-Hub-Signature-256"
)

func WebhookHandler(ctx context.Context, _ http.ResponseWriter, r listeners.RequestData) int {
	l := logger.FromContext(ctx).With(slog.String("link_type", "github"), slog.String("link_medium", "webhook"))
	t := time.Now().UTC()

	if statusCode := checkContentTypeHeader(l, r); statusCode != http.StatusOK {
		return otel.IncrementWebhookEventCounter(l, t, "", statusCode)
	}
	if statusCode := CheckSignatureHeader(l, r); statusCode != http.StatusOK {
		return otel.IncrementWebhookEventCounter(l, t, "", statusCode)
	}

	// If the payload is a web form, convert it to JSON.
	if r.Headers.Get(contentTypeHeader) == client.ContentForm {
		reader := strings.NewReader(r.WebForm.Get("payload"))
		if err := json.NewDecoder(reader).Decode(&r.JSONPayload); err != nil {
			l.Error("failed to extract and decode JSON payload from form data", slog.Any("error", err))
			return otel.IncrementWebhookEventCounter(l, t, "", http.StatusInternalServerError)
		}
	}

	// Dispatch the event notification as a Temporal signal.
	signalName := "github.events." + r.Headers.Get(eventHeader)
	if err := temporal.Signal(ctx, r.Temporal, signalName, r.JSONPayload); err != nil {
		l.Error("failed to send Temporal signal", slog.Any("error", err))
		return otel.IncrementWebhookEventCounter(l, t, signalName, http.StatusInternalServerError)
	}

	return otel.IncrementWebhookEventCounter(l, t, signalName, http.StatusOK)
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

// CheckSignatureHeader is defined by and for GitHub, but also reused by Bitbucket.
func CheckSignatureHeader(l *slog.Logger, r listeners.RequestData) int {
	sig := r.Headers.Get(signatureHeader)
	if sig == "" {
		l.Warn("bad request: missing header", slog.String("header", signatureHeader))
		return http.StatusForbidden
	}

	secret := r.LinkSecrets["webhook_secret"]
	if secret == "" {
		l.Warn("webhook secret is not configured")
		return http.StatusInternalServerError
	}

	if !verifySignature(l, secret, sig, r.RawPayload) {
		l.Warn("signature verification failed", slog.String("signature", sig),
			slog.Bool("has_signing_secret", secret != ""))
		return http.StatusForbidden
	}

	return http.StatusOK
}

// verifySignature implements
// https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries.
func verifySignature(l *slog.Logger, webhookSecret, want string, body []byte) bool {
	mac := hmac.New(sha256.New, []byte(webhookSecret))

	n, err := mac.Write(body)
	if err != nil {
		l.Error("HMAC write error", slog.Any("error", err))
		return false
	}
	if n != len(body) {
		return false
	}

	got := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(got), []byte(want))
}
