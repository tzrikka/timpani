package websocket

import (
	"bufio"
	"io"
	"log/slog"
	"net/http"
	"sync"
)

// Conn respresents the configuration and state of
// an open client connection to a WebSocket server.
type Conn struct {
	// Initialized before the handshake.
	logger  *slog.Logger
	client  *http.Client
	headers http.Header

	// Initialized after the handshake.
	bufio  *bufio.ReadWriter
	reader chan Message
	writer chan internalMessage
	closer io.ReadWriteCloser

	// No need for synchronization: value changes are possible only in
	// one direction (false to true), and are always done by a single
	// function, which is guaranteed to run in a single goroutine.
	closeReceived bool

	closeSent   bool
	closeSentMu sync.RWMutex

	// Only for the purpose of minimizing memory allocations (safely),
	// not for state management or memory sharing of any kind.
	readBuf  [8]byte
	writeBuf [8]byte
	closeBuf [maxControlPayload]byte

	// For unit-testing only.
	nonceGen io.Reader
}

// Message with WebSocket data, from one or more (defragmented) data frames,
// as defined in https://datatracker.ietf.org/doc/html/rfc6455#section-5.6.
// Returned by the Go channel that is exposed by [Conn.IncomingMessages].
type Message struct {
	Opcode Opcode
	Data   []byte
}

// internalMessage is used to synchronize concurrent calls to [Conn.writeFrame].
type internalMessage struct {
	Opcode Opcode
	Data   []byte
	err    chan<- error
}

// IncomingMessages returns the connection's channel that publishes
// data [Message]s as they are received from the server.
//
// [Message]: https://pkg.go.dev/github.com/tzrikka/timpani/pkg/websocket#Message
func (c *Conn) IncomingMessages() <-chan Message {
	return c.reader
}

// readMessages runs as a [Conn] goroutine, to call [Conn.readMessage]
// continuously, in order to process control and data frames, and
// publish data [Message]s to the connection's subscribers.
func (c *Conn) readMessages() {
	msg := c.readMessage()
	for msg != nil {
		c.reader <- Message{Opcode: msg.Opcode, Data: msg.Data}
		msg = c.readMessage()
	}
	close(c.reader)
}

// writeMessages runs as a [Conn] goroutine, to synchronize concurrent
// calls to [Conn.writeFrame]. For the time being, this package doesn't
// need to implement frame fragmentation in outbound messages.
func (c *Conn) writeMessages() {
	for msg := range c.writer {
		msg.err <- c.writeFrame(msg.Opcode, msg.Data)
		// The message's error channel can be used at most once.
		close(msg.err)
	}
}
