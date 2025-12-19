package jira

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/tzrikka/timpani/internal/listeners"
	"github.com/tzrikka/timpani/internal/logger"
	"github.com/tzrikka/timpani/pkg/metrics"
)

const (
	contentTypeHeader = "Content-Type"
	contentTypeJSON   = "application/json"
)

func WebhookHandler(ctx context.Context, _ http.ResponseWriter, r listeners.RequestData) int {
	l := logger.FromContext(ctx).With(slog.String("link_type", "jira"), slog.String("link_medium", "webhook"))
	t := time.Now().UTC()

	if ct := r.Headers.Get(contentTypeHeader); ct != contentTypeJSON {
		l.Warn("bad request: unexpected header value", slog.String("header", contentTypeHeader),
			slog.String("got", ct), slog.String("want", contentTypeJSON))
		return metrics.IncrementWebhookEventCounter(l, t, "", http.StatusBadRequest)
	}

	l.Warn("received Jira webhook event - processing not implemented yet")

	return metrics.IncrementWebhookEventCounter(l, t, "", http.StatusOK)
}
