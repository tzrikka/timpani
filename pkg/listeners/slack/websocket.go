package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/tzrikka/timpani/internal/listeners"
	"github.com/tzrikka/timpani/pkg/websocket"
)

const (
	connOpenURL = "https://slack.com/api/apps.connections.open"
	timeout     = 3 * time.Second
	maxSize     = 1024 // 1 KiB.
)

func ConnectionHandler(ctx context.Context, data listeners.LinkData) error {
	l := log.Logger.With().Str("link_type", "slack").Str("link_medium", "websocket").Logger()
	t := data.Secrets["app_token"]
	if t == "" {
		l.Warn().Msg("Thrippy link missing required credentials")
		return errors.New("forbidden")
	}

	c, err := websocket.NewOrCachedClient(ctx, urlFunc(t), t)
	if err != nil {
		l.Err(err).Msg("Slack Socket Mode connection error")
		return errors.New("internal server error")
	}

	go clientEventLoop(l, c)
	return nil
}

func urlFunc(appToken string) func(ctx context.Context) (string, error) {
	return func(ctx context.Context) (string, error) {
		return generateWebSocketURL(ctx, appToken)
	}
}

// generateWebSocketURL generates a temporary Socket Mode WebSocket URL ("wss://...")
// that an unpublished Slack app can connect to, to receive events and interactive
// payloads. Based on https://docs.slack.dev/reference/methods/apps.connections.open.
func generateWebSocketURL(ctx context.Context, appToken string) (string, error) {
	// Construct and send the request.
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, connOpenURL, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("failed to construct HTTP request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+appToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read and parse the response.
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxSize))
	if err != nil {
		return "", fmt.Errorf("failed to read HTTP response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		msg := resp.Status
		if len(body) > 0 {
			msg = fmt.Sprintf("%s: %s", msg, string(body))
		}
		return "", errors.New(msg)
	}

	decoded := &apiResponse{}
	if err := json.Unmarshal(body, decoded); err != nil {
		return "", fmt.Errorf("failed to parse JSON in HTTP response body: %w", err)
	}
	if !decoded.OK {
		return "", fmt.Errorf("error reported by Slack API: %s", decoded.Error)
	}

	return decoded.URL, nil
}

type apiResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
	URL   string `json:"url,omitempty"`
}

// clientEventLoop runs as a goroutine to parse, acknowledge, and dispatch
// all types of asynchronous Slack events which were received as WebSocket
// data messages. It also prevents downtime by informing the client when
// to refresh its underlying WebSocket connection, before it times out.
func clientEventLoop(l zerolog.Logger, c *websocket.Client) {
	for {
		raw, ok := <-c.IncomingMessages()
		if !ok {
			l.Error().Msg("WebSocket client is closed")
			return
		}

		msg := socketModeMessage{}
		if err := json.Unmarshal(raw.Data, &msg); err != nil {
			l.Err(err).Msg("JSON decoding error in incoming WebSocket message")
			continue
		}

		resp := eventResponse{EnvelopeID: msg.EnvelopeID}
		switch msg.Type {
		// https://docs.slack.dev/apis/events-api/using-socket-mode#connect
		case "hello":
			t := msg.DebugInfo.ApproximateConnectionTime
			// 63-72 seconds before the actual timeout.
			t -= 63 + rand.IntN(10) //gosec:disable G404 -- no need for crypto/rand
			c.RefreshConnectionIn(time.Duration(t) * time.Second)
			continue

		// https://docs.slack.dev/apis/events-api/using-socket-mode#disconnect
		case "disconnect":
			continue

		// https://docs.slack.dev/apis/events-api/using-socket-mode#command
		case "slash_commands":
			resp.Payload = map[string]any{
				"blocks": []map[string]any{
					{
						"type": "section",
						"text": map[string]string{
							"type": "mrkdwn",
							"text": fmt.Sprintf("Your command: `%s %s`", msg.Payload["command"], msg.Payload["text"]),
						},
					},
				},
			}
		}

		// https://docs.slack.dev/apis/events-api/using-socket-mode#acknowledge
		if err := c.SendJSONMessage(resp); err != nil {
			l.Err(err).Msg("failed to ack Slack Socket Mode event")
		}
	}
}

// https://docs.slack.dev/apis/events-api/using-socket-mode
type socketModeMessage struct {
	Type string `json:"type"`

	// Hello.
	NumConnections int           `json:"num_connections"`
	ConnectionInfo helloConnInfo `json:"connection_info"`

	// Disconnect.
	Reason string `json:"reason"`

	// Hello & disconnect.
	DebugInfo debugInfo `json:"debug_info"`

	// Events.
	Payload                map[string]any `json:"payload"`
	EnvelopeID             string         `json:"envelope_id"`
	AcceptsResponsePayload bool           `json:"accepts_response_payload"`
}

type helloConnInfo struct {
	AppID string `json:"app_id"`
}

type debugInfo struct {
	Host                      string `json:"host"`
	BuildNumber               int    `json:"build_number,omitempty"`
	ApproximateConnectionTime int    `json:"approximate_connection_time,omitempty"`
}

// https://docs.slack.dev/apis/events-api/using-socket-mode#acknowledge
type eventResponse struct {
	EnvelopeID string         `json:"envelope_id"`
	Payload    map[string]any `json:"payload,omitempty"`
}
