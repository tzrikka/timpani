package bitbucket

import (
	"context"
	"net/http"
	"strings"

	"github.com/rs/zerolog"

	"github.com/tzrikka/timpani/internal/listeners"
	"github.com/tzrikka/timpani/pkg/listeners/github"
	"github.com/tzrikka/timpani/pkg/temporal"
)

const (
	contentTypeHeader = "Content-Type"
	contentTypeJSON   = "application/json"
	eventHeader       = "X-Event-Key"
)

func WebhookHandler(ctx context.Context, _ http.ResponseWriter, r listeners.RequestData) int {
	l := zerolog.Ctx(ctx).With().Str("link_type", "bitbucket").Str("link_medium", "webhook").Logger()

	if ct := r.Headers.Get(contentTypeHeader); ct != contentTypeJSON {
		l.Warn().Str("header", contentTypeHeader).Str("got", ct).Any("want", contentTypeJSON).
			Msg("bad request: unexpected header value")
		return http.StatusBadRequest
	}

	if statusCode := github.CheckSignatureHeader(l, r); statusCode != http.StatusOK {
		return statusCode
	}

	// Dispatch the event notification as a Temporal signal.
	signalName := "bitbucket.events." + strings.ReplaceAll(r.Headers.Get(eventHeader), ":", ".")
	if err := temporal.Signal(ctx, r.Temporal, signalName, r.JSONPayload); err != nil {
		l.Err(err).Msg("failed to send Temporal signal")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}
