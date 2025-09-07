package websocket

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
)

// Opcode denotes the type of a WebSocket frame, as defined in
// https://datatracker.ietf.org/doc/html/rfc6455#section-5.2 and
// https://datatracker.ietf.org/doc/html/rfc6455#section-11.8.
type Opcode int

const (
	opcodeContinuation Opcode = iota
	OpcodeText
	OpcodeBinary
	// 3-7 are reserved for further non-control frames.
	_
	_
	_
	_
	_
	opcodeClose
	opcodePing
	opcodePong
	// 11-16 are reserved for further control frames.
)

// String returns the opcode's name, or its number if it's unrecognized.
func (o Opcode) String() string {
	switch o {
	case opcodeContinuation:
		return "continuation"
	case OpcodeText:
		return "text"
	case OpcodeBinary:
		return "binary"
	case opcodeClose:
		return "close"
	case opcodePing:
		return "ping"
	case opcodePong:
		return "pong"
	default:
		return strconv.Itoa(int(o))
	}
}

// Frame parsing/construction constants, as defined in
// https://datatracker.ietf.org/doc/html/rfc6455#section-5.2.
const (
	bit0     = 0x80
	bit1     = 0x40
	bit2     = 0x20
	bit3     = 0x10
	bits1to7 = 0x7f
	bits4to7 = 0x0f

	len7bits  = 125 // Payload length of up to 125 bytes.
	len16bits = 126 // Extended payload length of up to 64 KiB.
	len64bits = 127 // Extended payload length of up to 16 EiB.
)

// frameHeader is based on https://datatracker.ietf.org/doc/html/rfc6455#section-5.2,
// excluding the masking key and payload data.
type frameHeader struct {
	// Bit 0: Indicates that this is the final fragment in a message.
	// The first fragment MAY also be the final fragment.
	fin bool
	// Bits 1-3: Reserved.
	rsv [3]bool
	// Bits 4-7: Defines the interpretation of the "Payload data".
	opcode Opcode
	// Bit 8: Defines whether the "Payload data" is masked. If set to 1, a masking key
	// is present in masking-key, and this is used to unmask the "Payload data" as per
	// [Section 5.3]. All frames sent from client to server have this bit set to 1.
	//
	// [Section 5.3]: https://datatracker.ietf.org/doc/html/rfc6455#section-5.3
	mask bool
	// Bits 9-15 + 0 or 2 or 8 bytes: The length of the "Payload data", in bytes: if
	// 0-125, that is the payload length. If 126, the following 2 bytes interpreted as
	// a 16-bit unsigned integer are the payload length. If 127, the following 8 bytes
	// interpreted as a 64-bit unsigned integer (the most significant bit MUST be 0) are
	// the payload length. Multibyte length quantities are expressed in network byte
	// order. Note that in all cases, the minimal number of bytes MUST be used to encode
	// the length, for example, the length of a 124-byte-long string can't be encoded as
	// the sequence 126, 0, 124. The payload length is the length of the "Extension data"
	// + the length of the "Application data". The length of the "Extension data" may be
	// zero, in which case the payload length is the length of the "Application data".
	payloadLength uint64
}

// readFrameHeader reads a frame received from the server,
// except for the payload. It blocks until such a frame exists.
//
// It is based on:
//   - Base framing protocol: https://datatracker.ietf.org/doc/html/rfc6455#section-5.2
//   - Receiving data: https://datatracker.ietf.org/doc/html/rfc6455#section-6.2
func (c *Conn) readFrameHeader() (frameHeader, error) {
	h := frameHeader{}

	// (Wait for and) read the first byte.
	b, err := c.bufio.ReadByte()
	if err != nil {
		return h, fmt.Errorf("failed to read first byte of incoming WebSocket frame: %w", err)
	}

	h.fin = (b & bit0) != 0
	h.rsv[0] = (b & bit1) != 0
	h.rsv[1] = (b & bit2) != 0
	h.rsv[2] = (b & bit3) != 0
	h.opcode = Opcode(b & bits4to7)

	// Read the second byte.
	b, err = c.bufio.ReadByte()
	if err != nil {
		return h, fmt.Errorf("failed to read second byte of incoming WebSocket frame: %w", err)
	}

	h.mask = (b & bit0) != 0

	b &= bits1to7
	switch {
	case b <= len7bits:
		h.payloadLength = uint64(b)
	case b == len16bits:
		_, err = io.ReadFull(c.bufio, c.readBuf[:2])
		h.payloadLength = uint64(binary.BigEndian.Uint16(c.readBuf[:2]))
	case b == len64bits:
		_, err = io.ReadFull(c.bufio, c.readBuf[:8])
		h.payloadLength = binary.BigEndian.Uint64(c.readBuf[:8])
	}
	if err != nil {
		return h, fmt.Errorf("failed to read payload length of incoming WebSocket frame: %w", err)
	}

	return h, nil
}

// maxControlPayload is the maximum length of a control frame payload,
// as defined in https://datatracker.ietf.org/doc/html/rfc6455#section-5.5.
const (
	maxControlPayload = 125
)

// checkFrameHeader checks if the connection needs to be closed, in case the
// server sent an invalid frame. If so, it also returns a human-readable reason.
//
// It is based on:
//   - Overview: https://datatracker.ietf.org/doc/html/rfc6455#section-5.1
//   - Base framing protocol: https://datatracker.ietf.org/doc/html/rfc6455#section-5.2
//   - Control frames: https://datatracker.ietf.org/doc/html/rfc6455#section-5.5
func (c *Conn) checkFrameHeader(h frameHeader, msgType Opcode) (string, error) {
	// "Reserved bits MUST be 0 unless an extension is negotiated that defines
	// meanings for non-zero values. If a nonzero value is received and none of
	// the negotiated extensions defines the meaning of such a nonzero value,
	// the receiving endpoint MUST _Fail the WebSocket Connection_".
	if h.rsv[0] || h.rsv[1] || h.rsv[2] {
		reason := "invalid reserved bits"
		return reason, fmt.Errorf("WebSocket server sent %s", reason)
	}

	// "If an unknown opcode is received, the receiving
	// endpoint MUST _Fail the WebSocket Connection_".
	if (h.opcode > 2 && h.opcode < 8) || h.opcode > 10 {
		reason := fmt.Sprintf("unknown opcode %d", h.opcode)
		return reason, fmt.Errorf("WebSocket server sent %s", reason)
	}

	// "A fragmented message consists of a single frame with the FIN bit
	// clear and an opcode other than 0, followed by zero or more frames
	// with the FIN bit clear and the opcode set to 0, and terminated by
	// a single frame with the FIN bit set and an opcode of 0".
	if h.opcode == opcodeContinuation && msgType == opcodeContinuation {
		reason := "continuation frame with nothing to continue"
		return reason, fmt.Errorf("WebSocket server sent %s", reason)
	}
	if (h.opcode == OpcodeText || h.opcode == OpcodeBinary) && msgType != opcodeContinuation {
		reason := "continuation frame with non-continuation opcode"
		return reason, fmt.Errorf("WebSocket server sent %s", reason)
	}

	// "All control frames MUST have a payload length of
	// 125 bytes or less and MUST NOT be fragmented".
	if h.opcode > 7 {
		if h.payloadLength > maxControlPayload {
			reason := "payload length too big"
			return reason, fmt.Errorf("WebSocket control frame (opcode %d) too large: %d bytes", h.opcode, h.payloadLength)
		}
		if !h.fin {
			reason := "control frame must not be fragmented"
			return reason, fmt.Errorf("WebSocket control frame (opcode %d) must not be fragmented", h.opcode)
		}
	}

	// "A server MUST NOT mask any frames that it sends to the client.
	// A client MUST close a connection if it detects a masked frame".
	if h.mask {
		reason := "server payloads must not be masked"
		return reason, errors.New("WebSocket server masked the payload data")
	}

	return "", nil
}

// writeFrame is optimized to send a single, unfragmented, masked frame.
//
// Do not call this function directly, call [sendControlFrame] instead,
// to ensure we always send one frame at a time!
//
// This function is based on:
//   - Base framing protocol: https://datatracker.ietf.org/doc/html/rfc6455#section-5.2
//   - Client-to-server masking: https://datatracker.ietf.org/doc/html/rfc6455#section-5.3
//   - Sending data: https://datatracker.ietf.org/doc/html/rfc6455#section-6.1
func (c *Conn) writeFrame(op Opcode, payload []byte) error {
	// Construct the header (automatically set the FIN and MASKED bits).
	if err := c.bufio.WriteByte(bit0 | byte(op)); err != nil {
		return fmt.Errorf("failed to write WebSocket control frame header: %w", err)
	}

	if err := c.writePayloadLength(len(payload)); err != nil {
		return fmt.Errorf("failed to write WebSocket control frame header: %w", err)
	}

	// Generate a random client masking key.
	if _, err := io.ReadFull(rand.Reader, c.writeBuf[:4]); err != nil {
		return fmt.Errorf("failed to generate masking key for WebSocket client frame: %w", err)
	}

	if _, err := c.bufio.Write(c.writeBuf[:4]); err != nil {
		return fmt.Errorf("failed to write WebSocket control frame masking key: %w", err)
	}

	// Mask and copy the payload.
	if len(payload) > 0 {
		c.mask(payload)
		defer c.mask(payload) // Undo the masking before returning.

		if _, err := c.bufio.Write(payload); err != nil {
			return fmt.Errorf("failed to write WebSocket control frame payload: %w", err)
		}
	}

	// Send the frame to the server.
	if err := c.bufio.Flush(); err != nil {
		return fmt.Errorf("failed to flush after writing WebSocket control frame: %w", err)
	}

	return nil
}

// writePayloadLength implements the payload length formatting which is
// defined in https://datatracker.ietf.org/doc/html/rfc6455#section-5.2.
func (c *Conn) writePayloadLength(n int) error {
	switch {
	// Up to 125 bytes (0 extra bytes).
	case n <= maxControlPayload:
		return c.bufio.WriteByte(bit0 | byte(n))

	// Up to 64 KiB (2 extra bytes).
	case n <= math.MaxUint16:
		if err := c.bufio.WriteByte(bit0 | len16bits); err != nil {
			return err
		}
		binary.BigEndian.PutUint16(c.writeBuf[:2], uint16(n)) //gosec:disable G115 -- value checked before cast
		_, err := c.bufio.Write(c.writeBuf[:2])
		return err

	// Up to 16 EiB (8 extra bytes).
	default:
		if err := c.bufio.WriteByte(bit0 | len64bits); err != nil {
			return err
		}
		binary.BigEndian.PutUint64(c.writeBuf[:8], uint64(n)) //gosec:disable G115 -- value checked before cast
		_, err := c.bufio.Write(c.writeBuf[:8])
		return err
	}
}

// mask implements https://datatracker.ietf.org/doc/html/rfc6455#section-5.3.
// Notice that it changes the input slice in-place! However, this function
// is its own inverse: applying it twice on the same payload
// results in the original unmasked payload.
func (c *Conn) mask(payload []byte) {
	for i := range payload {
		payload[i] ^= c.writeBuf[i&3]
	}
}
