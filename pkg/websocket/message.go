package websocket

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"unicode/utf8"
)

// readMessage reads incoming frames from the server, responds to
// control frames (whether or not they're interleaved with data frames),
// and defragments data frames if needed. This function handles errors
// and connection closures gracefully, and returns nil in such cases.
//
// Do not call this function directly, it is meant to be used
// exclusively (and continuously) by [Conn.readMessages]!
//
// It is based on:
//   - Base framing protocol: https://datatracker.ietf.org/doc/html/rfc6455#section-5.2
//   - Fragmentation: https://datatracker.ietf.org/doc/html/rfc6455#section-5.4
//   - Control frames: https://datatracker.ietf.org/doc/html/rfc6455#section-5.5
//   - Data frames: https://datatracker.ietf.org/doc/html/rfc6455#section-5.6
//   - Receiving data: https://datatracker.ietf.org/doc/html/rfc6455#section-6.2
//   - Closing the connection: https://datatracker.ietf.org/doc/html/rfc6455#section-7
//   - Handling Errors in UTF-8-Encoded Data: https://datatracker.ietf.org/doc/html/rfc6455#section-8.1
func (c *Conn) readMessage() *internalMessage {
	var msg bytes.Buffer
	var op Opcode

	for {
		h, err := c.readFrameHeader()
		if err != nil {
			if errors.Is(err, io.EOF) {
				c.logger.Debug("WebSocket connection closed")
				c.closeReceived = true
				c.closeSent = true
				return nil
			}
			c.logger.Error("failed to read WebSocket frame header", slog.Any("error", err))
			c.sendCloseControlFrame(StatusInternalError, "frame header reading error")
			return nil
		}

		c.logger.Debug("received WebSocket frame", slog.Bool("fin", h.fin),
			slog.String("opcode", h.opcode.String()), slog.Any("length", h.payloadLength))

		var data []byte
		if h.payloadLength > 0 {
			data = make([]byte, h.payloadLength)
			if _, err := io.ReadFull(c.bufio, data); err != nil {
				c.logger.Error("failed to read WebSocket frame payload", slog.Any("error", err))
				c.sendCloseControlFrame(StatusInternalError, "frame payload reading error")
				return nil
			}
		}

		if reason, err := c.checkFrameHeader(h, op); err != nil {
			c.logger.Error("protocol error due to invalid frame", slog.Any("error", err))
			c.sendCloseControlFrame(StatusProtocolError, reason)
			return nil
		}

		switch h.opcode {
		// "A fragmented message consists of a single frame with the FIN bit
		// clear and an opcode other than 0, followed by zero or more frames
		// with the FIN bit clear and the opcode set to 0, and terminated by
		// a single frame with the FIN bit set and an opcode of 0".
		case opcodeContinuation, OpcodeText, OpcodeBinary:
			if h.opcode != opcodeContinuation {
				op = h.opcode
			}
			if h.payloadLength > 0 {
				if _, err := msg.Write(data); err != nil {
					c.logger.Error("failed to store WebSocket data frame payload", slog.Any("error", err))
					c.sendCloseControlFrame(StatusInternalError, "data frame payload storing error")
					return nil
				}
			}

		// "If an endpoint receives a Close frame and did not previously send
		// a Close frame, the endpoint MUST send a Close frame in response".
		case opcodeClose:
			c.closeReceived = true
			status, reason := c.parseClosePayload(data)
			c.sendCloseControlFrame(status, reason)
			return nil // Not an error, but we no longer need to receive new frames.

		// "An endpoint MUST be capable of handling control
		// frames in the middle of a fragmented message".
		case opcodePing:
			if err := <-c.sendControlFrame(opcodePong, data); err != nil {
				c.logger.Error("failed to send WebSocket pong control frame",
					slog.Any("error", err), slog.Any("payload", data))
			}

		case opcodePong:
			// No need to handle "Pong" control frames, since this
			// client doesn't send unsolicited "Ping" control frames.
		}

		if h.fin && h.opcode <= OpcodeBinary {
			return c.finalizeMessage(op, msg.Bytes())
		}
	}
}

func (c *Conn) finalizeMessage(op Opcode, data []byte) *internalMessage {
	if data == nil {
		data = []byte{}
	}

	c.logger.Debug("finished receiving WebSocket data message",
		slog.String("opcode", op.String()), slog.Int("length", len(data)))

	// "When an endpoint is to interpret a byte stream as UTF-8 but finds
	// that the byte stream is not, in fact, a valid UTF-8 stream, that
	// endpoint MUST _Fail the WebSocket Connection_. This rule applies both
	// during the opening handshake and during subsequent data exchange".
	if op == OpcodeText && len(data) > 0 && !utf8.Valid(data) {
		c.logger.Error("protocol error due to invalid UTF-8 text")
		c.sendCloseControlFrame(StatusInvalidData, "invalid UTF-8 text")
		return nil
	}

	return &internalMessage{Opcode: op, Data: data}
}

// SendTextMessage sends a [UTF-8 text] message to the server.
//
// This is done asynchronously, to manage [isolation or safe multiplexing]
// of multiple concurrent calls, including interleaved control frames.
// Despite that, this function enables the caller to block and/or
// handle errors, with the returned channel.
//
// [UTF-8 text]: https://datatracker.ietf.org/doc/html/rfc6455#section-5.6
// [isolation or safe multiplexing]: https://datatracker.ietf.org/doc/html/rfc6455#section-5.4
func (c *Conn) SendTextMessage(data []byte) <-chan error {
	err := make(chan error)
	c.writer <- internalMessage{Opcode: OpcodeText, Data: data, err: err}
	return err
}

// SendBinaryMessage sends a [binary] message to the server.
//
// This is done asynchronously, to manage [isolation or safe multiplexing]
// of multiple concurrent calls, including interleaved control frames.
// Despite that, this function enables the caller to block and/or
// handle errors, with the returned channel.
//
// [binary]: https://datatracker.ietf.org/doc/html/rfc6455#section-5.6
// [isolation or safe multiplexing]: https://datatracker.ietf.org/doc/html/rfc6455#section-5.4
func (c *Conn) SendBinaryMessage(data []byte) <-chan error {
	err := make(chan error)
	c.writer <- internalMessage{Opcode: OpcodeBinary, Data: data, err: err}
	return err
}

// sendControlFrame sends a [WebSocket control frame] to the server.
//
// This is done asynchronously, to manage [isolation or safe multiplexing]
// of multiple concurrent calls, including interleaved control frames.
// Despite that, this function enables the caller to block and/or
// handle errors, with the returned channel.
//
// Use this function instead of calling [writeFrame] directly!
//
// [WebSocket control frame]: https://datatracker.ietf.org/doc/html/rfc6455#section-5.5
func (c *Conn) sendControlFrame(op Opcode, payload []byte) <-chan error {
	err := make(chan error)
	c.writer <- internalMessage{Opcode: op, Data: payload, err: err}
	return err
}
