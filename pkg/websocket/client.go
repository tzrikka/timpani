package websocket

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/tzrikka/timpani/internal/logger"
)

var clients = sync.Map{}

// Client is a long-running wrapper of connections to the same WebSocket
// server with the same credentials. It usually manages a single [Conn],
// except when it gets disconnected, or is about to be, in which case the
// client automatically opens another [Conn] and switches to it seamlessly,
// to prevent or at least minimize downtime during reconnections.
type Client struct {
	logger *slog.Logger
	url    urlFunc
	opts   []DialOpt

	conns   [2]*Conn
	inMsgs  <-chan Message
	outMsgs chan Message

	refresh *time.Timer
}

type urlFunc func(ctx context.Context) (string, error)

func NewOrCachedClient(ctx context.Context, url urlFunc, id string, opts ...DialOpt) (*Client, error) {
	hashedID := hash(id)
	if client, ok := clients.Load(hashedID); ok {
		return client.(*Client), nil //nolint:errcheck
	}

	c, err := newClient(ctx, url, opts...)
	if err != nil {
		return nil, err
	}

	actual, loaded := clients.LoadOrStore(hashedID, c)
	if loaded { // Stored by a different goroutine since clients.Load() above.
		deleteClient(c)
	} else { // Newly-stored by this goroutine, so activate its message relay.
		go c.relayMessages(ctx)
	}

	return actual.(*Client), nil //nolint:errcheck
}

// hash generates a stable-but-irreversible SHA-256 hash of a [Client] ID.
func hash(id string) string {
	h := sha256.New()
	h.Write([]byte(id))
	return hex.EncodeToString(h.Sum(nil))
}

func newClient(ctx context.Context, f urlFunc, opts ...DialOpt) (*Client, error) {
	conn, err := newConn(ctx, f, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		logger:  logger.FromContext(ctx),
		url:     f,
		opts:    opts,
		conns:   [2]*Conn{conn},
		inMsgs:  conn.IncomingMessages(),
		outMsgs: make(chan Message),
	}, nil
}

func newConn(ctx context.Context, f urlFunc, opts ...DialOpt) (*Conn, error) {
	url, err := f(ctx)
	if err != nil {
		return nil, err
	}

	return Dial(ctx, url, opts...)
}

func (c *Client) newConn(ctx context.Context, f urlFunc, opts ...DialOpt) (*Conn, error) {
	return newConn(logger.WithContext(ctx, c.logger), f, opts...)
}

// deleteClient deletes a newly-created [Client] which is not needed anymore,
// because a different one was already activated with the same ID.
func deleteClient(c *Client) {
	c.conns[0].Close(StatusGoingAway)

	c.logger = nil
	c.url = nil
	c.opts = nil

	c.conns = [2]*Conn{}
	c.inMsgs = nil
	c.outMsgs = nil
}

// relayMessages runs as a [Client] goroutine, to route data [Message]s
// from the client's underlying [Conn] to the client's subscribers.
func (c *Client) relayMessages(ctx context.Context) {
	for {
		if msg, ok := <-c.inMsgs; ok {
			c.outMsgs <- msg
			continue
		}

		c.replaceConn(ctx)
	}
}

// replaceConn either creates a new [Conn] (if the existing one is
// closing/closed), or switches seamlessly to a secondary one which
// was created by the timer-based goroutine in [RefreshConnectionIn].
func (c *Client) replaceConn(ctx context.Context) {
	defer func() {
		c.inMsgs = c.conns[0].IncomingMessages()
	}()

	// Switch to a fresh secondary connection.
	if c.conns[1] != nil {
		c.conns[0] = c.conns[1]
		c.conns[1] = nil
		return
	}

	// Create a new connection, with endless retries.
	i := 0
	for {
		conn, err := c.newConn(ctx, c.url, c.opts...)
		if err == nil {
			c.conns[0] = conn
			break
		}

		c.logger.Error("failed to replace WebSocket connection", slog.Any("error", err), slog.Int("retry", i))
		i++
	}
}

// IncomingMessages returns the client's channel that publishes
// data [Message]s as they are received from the server.
//
// [Message]: https://pkg.go.dev/github.com/tzrikka/timpani/pkg/websocket#Message
func (c *Client) IncomingMessages() <-chan Message {
	return c.outMsgs
}

// RefreshConnectionIn instructs the client to replace its underlying [Conn]
// seamlessly after the given duration of time. This prevents unnecessary
// downtime during normal reconnections, which is useful in connections
// where the disconnection time is known or coordinated in advance.
func (c *Client) RefreshConnectionIn(ctx context.Context, d time.Duration) {
	m := "starting timer to refresh WebSocket connection"
	if c.refresh != nil {
		c.refresh.Stop()
		m = "re" + m
	}
	c.logger.Debug(m)

	c.refresh = time.AfterFunc(d, func() {
		c.logger.Debug("refreshing WebSocket connection")
		c.refresh = nil

		conn, err := c.newConn(ctx, c.url, c.opts...)
		if err != nil {
			c.logger.Error("failed to refresh WebSocket connection", slog.Any("error", err))
			return
		}

		c.conns[1] = conn
		c.conns[0].Close(StatusGoingAway)
	})
}

// SendJSONMessage sends a JSON text message to the server.
func (c *Client) SendJSONMessage(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return <-c.conns[0].SendTextMessage(b)
}
