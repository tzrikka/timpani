package jira

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/tzrikka/timpani/internal/listeners"
	"github.com/tzrikka/timpani/pkg/metrics"
)

const (
	contentTypeHeader = "Content-Type"
	contentTypeJSON   = "application/json"
)

func WebhookHandler(ctx context.Context, _ http.ResponseWriter, r listeners.RequestData) int {
	l := zerolog.Ctx(ctx).With().Str("link_type", "jira").Str("link_medium", "webhook").Logger()
	t := time.Now().UTC()

	if ct := r.Headers.Get(contentTypeHeader); ct != contentTypeJSON {
		l.Warn().Str("header", contentTypeHeader).Str("got", ct).Str("want", contentTypeJSON).
			Msg("bad request: unexpected header value")
		return metrics.CountWebhookEvent(t, "", http.StatusBadRequest)
	}

	l.Warn().Msg("received Jira webhook event - processing not implemented yet")

	return metrics.CountWebhookEvent(t, "", http.StatusOK)
}
