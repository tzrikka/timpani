package bitbucket

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/tzrikka/timpani/internal/listeners"
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
	l := zerolog.Ctx(ctx).With().Str("link_type", "bitbucket").Str("link_medium", "webhook").Logger()
	t := time.Now().UTC()

	if ct := r.Headers.Get(contentTypeHeader); ct != contentTypeJSON {
		l.Warn().Str("header", contentTypeHeader).Str("got", ct).Any("want", contentTypeJSON).
			Msg("bad request: unexpected header value")
		return metrics.CountWebhookEvent(t, "", http.StatusBadRequest)
	}

	// Note 1: Bitbucket uses the exact same signature checking method as GitHub.
	// Note 2: Some large customers of Bitbnucket use proxies to fan-out webhook
	// events instead of using many webhook registrations, in order to avoid
	// hitting rate limits. In such cases, the webhook secret may be blank.
	if r.LinkSecrets["webhook_secret"] != "" {
		if statusCode := github.CheckSignatureHeader(l, r); statusCode != http.StatusOK {
			return metrics.CountWebhookEvent(t, "", statusCode)
		}
	}

	// Dispatch the event notification as a Temporal signal.
	signalName := "bitbucket.events." + strings.ReplaceAll(r.Headers.Get(eventHeader), ":", ".")
	if err := temporal.Signal(ctx, r.Temporal, signalName, r.JSONPayload); err != nil {
		l.Err(err).Msg("failed to send Temporal signal")
		return metrics.CountWebhookEvent(t, signalName, http.StatusInternalServerError)
	}

	return metrics.CountWebhookEvent(t, signalName, http.StatusOK)
}
