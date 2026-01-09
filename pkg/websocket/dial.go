package websocket

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha1" //gosec:disable G505 // Required by the WebSocket protocol.
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/tzrikka/timpani/internal/logger"
)

type DialOpt func(*Conn)

var defaultClient = adjustHTTPClient(*http.DefaultClient)

// WithHTTPClient lets callers of [Dial] specify a custom [http.Client]
// to use for the WebSocket handshake, instead of [http.DefaultClient].
//
// Do not specify a custom timeout in the HTTP client! This will interfere with
// the long-lived WebSocket connection beyond the scope of its initial handshake.
// Instead, use [context.WithTimeout] with the [context.Context] passed to [Dial].
func WithHTTPClient(hc *http.Client) DialOpt {
	return func(c *Conn) {
		c.client = hc
	}
}

// WithHTTPHeader lets callers of [Dial] add a single HTTP header to the WebSocket
// handshake's HTTP request. Use [WithHTTPHeaders] to specify multiple ones.
func WithHTTPHeader(key, value string) DialOpt {
	return func(c *Conn) {
		c.headers.Add(key, value)
	}
}

// WithHTTPHeaders lets callers of [Dial] add multiple HTTP headers to the WebSocket
// handshake's HTTP request, instead of calling [WithHTTPHeader] multiple times.
func WithHTTPHeaders(hs http.Header) DialOpt {
	return func(c *Conn) {
		c.headers = hs.Clone()
	}
}

// Dial performs a [WebSocket handshake] to establish
// a connection to the given URL ("ws://..." or "wss://").
//
// [WebSocket handshake]: https://datatracker.ietf.org/doc/html/rfc6455#section-4.1
func Dial(ctx context.Context, wsURL string, opts ...DialOpt) (*Conn, error) {
	// Initialize optional configuration details and internal helpers.
	c := &Conn{
		logger:   logger.FromContext(ctx),
		headers:  http.Header{},
		nonceGen: rand.Reader,
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.client == nil {
		c.client = defaultClient
	} else {
		c.client = adjustHTTPClient(*c.client)
	}

	// Send handshake request & check response.
	nonce, err := generateNonce(c.nonceGen)
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce for WebSocket handshake: %w", err)
	}
	req, err := c.handshakeRequest(ctx, wsURL, nonce)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send WebSocket handshake request: %w", err)
	}
	if err = checkHandshakeResponse(resp, nonce); err != nil {
		_ = resp.Body.Close()
		return nil, err
	}

	// Post-handshake connection state initializations.
	rwc, ok := resp.Body.(io.ReadWriteCloser)
	if !ok {
		return nil, fmt.Errorf("WebSocket handshake response body type: got %T, want io.ReadWriteCloser", resp.Body)
	}

	c.bufio = bufio.NewReadWriter(bufio.NewReader(rwc), bufio.NewWriter(rwc))
	c.reader = make(chan Message)
	c.writer = make(chan internalMessage)
	c.closer = rwc

	go c.readMessages()
	go c.writeMessages()

	c.logger.Debug("WebSocket connection initialized")
	return c, nil
}

// adjustHTTPClient returns a modified shallow copy of the given [http.Client].
func adjustHTTPClient(c http.Client) *http.Client {
	// Wrap the HTTP client's CheckRedirect function, to convert
	// ws/wss URL schemes to http/https, respectively.
	origCheckRedirect := c.CheckRedirect
	c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		switch req.URL.Scheme {
		case "ws":
			req.URL.Scheme = "http"
		case "wss":
			req.URL.Scheme = "https"
		}

		if origCheckRedirect != nil {
			return origCheckRedirect(req, via)
		}
		return nil
	}

	return &c
}

// generateNonce generates a nonce consisting of a randomly
// selected 16-byte value that has been Base64-encoded. The
// nonce MUST be selected randomly for each connection.
func generateNonce(r io.Reader) (string, error) {
	b := make([]byte, 16)
	if _, err := io.ReadFull(r, b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// handshakeRequest implements the client request details
// in https://datatracker.ietf.org/doc/html/rfc6455#section-4.1.
func (c *Conn) handshakeRequest(ctx context.Context, wsURL, nonce string) (*http.Request, error) {
	u, err := url.Parse(wsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse WebSocket URL: %w", err)
	}

	switch u.Scheme {
	case "ws":
		u.Scheme = "http"
	case "wss":
		u.Scheme = "https"
	case "http", "https":
		// Do nothing.
	default:
		return nil, fmt.Errorf("unexpected WebSocket URL scheme: %q", u.Scheme)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebSocket handshake request: %w", err)
	}

	req.Header = c.headers.Clone()
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Key", nonce)
	req.Header.Set("Sec-WebSocket-Version", "13")
	// Sec-WebSocket-Extensions, Sec-WebSocket-Protocol.

	return req, nil
}

// checkHandshakeResponse checks the server response details in
// https://datatracker.ietf.org/doc/html/rfc6455#section-4.2.2.
func checkHandshakeResponse(resp *http.Response, nonce string) error {
	if resp.StatusCode != http.StatusSwitchingProtocols {
		msg := "WebSocket handshake response status: got %d, want %d"
		msg = fmt.Sprintf(msg, resp.StatusCode, http.StatusSwitchingProtocols)

		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		if len(body) > 0 {
			msg = fmt.Sprintf("%s (%s)", msg, string(body))
		}

		return errors.New(msg)

	}

	if err := checkHTTPHeader(resp.Header, "Upgrade", "websocket"); err != nil {
		return err
	}

	if err := checkHTTPHeader(resp.Header, "Connection", "Upgrade"); err != nil {
		return err
	}

	want := expectedServerAcceptValue(nonce)
	if err := checkHTTPHeader(resp.Header, "Sec-WebSocket-Accept", want); err != nil {
		return err
	}

	// Sec-WebSocket-Protocol, Sec-WebSocket-Extensions.

	return nil
}

func checkHTTPHeader(headers http.Header, key, want string) error {
	if got := headers.Get(key); !strings.EqualFold(got, want) {
		return fmt.Errorf("WebSocket handshake response header %q: got %q, want %q", key, got, want)
	}
	return nil
}

var acceptGUID = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")

// expectedServerAcceptValue constructs the expected value of the "Sec-WebSocket-Accept"
// header, as defined in https://datatracker.ietf.org/doc/html/rfc6455#section-4.2.2.
func expectedServerAcceptValue(key string) string {
	h := sha1.New() //gosec:disable G401 // Required by the WebSocket protocol.
	h.Write([]byte(key))
	h.Write(acceptGUID)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
