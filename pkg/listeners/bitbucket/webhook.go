package bitbucket

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/tzrikka/timpani/internal/listeners"
	"github.com/tzrikka/timpani/internal/logger"
	"github.com/tzrikka/timpani/pkg/listeners/github"
	"github.com/tzrikka/timpani/pkg/metrics"
	"github.com/tzrikka/timpani/pkg/temporal"
)

const (
	contentTypeHeader = "Content-Type"
	contentTypeJSON   = "application/json"
	eventHeader       = "X-Event-Key"
)

func WebhookHandler(ctx context.Context, _ http.ResponseWriter, r listeners.RequestData) int {
	l := logger.FromContext(ctx).With(slog.String("link_type", "bitbucket"), slog.String("link_medium", "webhook"))
	t := time.Now().UTC()

	if ct := r.Headers.Get(contentTypeHeader); ct != contentTypeJSON {
		l.Warn("bad request: unexpected header value", slog.String("header", contentTypeHeader),
			slog.String("got", ct), slog.String("want", contentTypeJSON))
		return metrics.IncrementWebhookEventCounter(l, t, "", http.StatusBadRequest)
	}

	// Note 1: Bitbucket uses the exact same signature checking method as GitHub.
	// Note 2: Some large customers of Bitbnucket use proxies to fan-out webhook
	// events instead of using many webhook registrations, in order to avoid
	// hitting rate limits. In such cases, the webhook secret may be blank.
	if r.LinkSecrets["webhook_secret"] != "" {
		if statusCode := github.CheckSignatureHeader(l, r); statusCode != http.StatusOK {
			return metrics.IncrementWebhookEventCounter(l, t, "", statusCode)
		}
	}

	// Dispatch the event notification as a Temporal signal.
	signalName := "bitbucket.events." + strings.ReplaceAll(r.Headers.Get(eventHeader), ":", ".")
	if err := temporal.Signal(ctx, r.Temporal, signalName, r.JSONPayload); err != nil {
		l.Error("failed to send Temporal signal", slog.Any("error", err))
		return metrics.IncrementWebhookEventCounter(l, t, signalName, http.StatusInternalServerError)
	}

	return metrics.IncrementWebhookEventCounter(l, t, signalName, http.StatusOK)
}
