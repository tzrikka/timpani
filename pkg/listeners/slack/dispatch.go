package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/rs/zerolog"

	"github.com/tzrikka/timpani/internal/listeners"
	"github.com/tzrikka/timpani/pkg/temporal"
)

func dispatchFromWebhook(ctx context.Context, r listeners.RequestData) error {
	l := zerolog.Ctx(ctx)

	signalName, payload, err := parsePayload(r.JSONPayload, r.WebForm)
	if err != nil {
		l.Err(err).Msg("failed to decode event payload")
		return err
	}

	if err := temporal.Signal(ctx, r.Temporal, signalName, payload); err != nil {
		l.Err(err).Msg("failed to send Temporal signal")
		return err
	}

	return nil
}

func dispatchFromWebSocket(l zerolog.Logger, tc listeners.TemporalConfig, payload map[string]any) error {
	signalName, payload, err := parsePayload(payload, nil)
	if err != nil {
		l.Err(err).Msg("failed to decode event payload")
		return err
	}

	ctx := l.WithContext(context.Background())
	if err := temporal.Signal(ctx, tc, signalName, payload); err != nil {
		l.Err(err).Msg("failed to send Temporal signal")
		return err
	}

	return nil
}

func parsePayload(payload map[string]any, webForm url.Values) (string, map[string]any, error) {
	// https://docs.slack.dev/apis/events-api#events-JSON
	eventType := payload["type"]
	if eventType == "event_callback" {
		if m, ok := payload["event"].(map[string]any); ok {
			eventType = m["type"]
		}
	}

	// https://docs.slack.dev/interactivity/implementing-slash-commands#app_command_handling
	if webForm.Get("command") != "" {
		eventType = "slash_command"
		payload = webFormToMap(webForm)
	}
	// https://docs.slack.dev/apis/events-api/using-socket-mode#command
	if payload["command"] != nil {
		eventType = "slash_command"
	}

	// https://docs.slack.dev/interactivity/handling-user-interaction#payloads
	// https://docs.slack.dev/reference/interaction-payloads
	if p := webForm.Get("payload"); p != "" {
		payload = map[string]any{}
		if err := json.NewDecoder(strings.NewReader(p)).Decode(&payload); err != nil {
			return "", nil, err
		}
		eventType = payload["type"]
	}

	return fmt.Sprintf("slack.events.%s", eventType), payload, nil
}

// webFormToMap converts a web form into a Go map which is compatible with JSON.
func webFormToMap(vs url.Values) map[string]any {
	m := make(map[string]any)
	for k, v := range vs {
		m[k] = v[0]
	}
	return m
}
