package websocket

import (
	"bufio"
	"bytes"
	"reflect"
	"testing"
)

// https://datatracker.ietf.org/doc/html/rfc6455#section-5.7
func TestConnReadFrameHeader(t *testing.T) {
	tests := []struct {
		name    string
		reader  []byte
		want    frameHeader
		wantErr bool
	}{
		{
			name:   "unmasked_text_hello",
			reader: []byte{0x81, 0x05, 0x48, 0x65, 0x6c, 0x6f},
			want:   frameHeader{fin: true, opcode: OpcodeText, payloadLength: 5},
		},
		{
			name:   "masked_text_hello",
			reader: []byte{0x81, 0x85, 0x37, 0xfa, 0x21, 0x3d, 0x7f, 0x9f, 0x4d, 0x51, 0x58},
			want:   frameHeader{fin: true, opcode: OpcodeText, mask: true, payloadLength: 5},
		},
		{
			name:   "first_fragment_unmasked_text_hel",
			reader: []byte{0x01, 0x03, 0x48, 0x65, 0x6c},
			want:   frameHeader{opcode: OpcodeText, payloadLength: 3},
		},
		{
			name:   "unmasked_ping",
			reader: []byte{0x89, 0x05, 0x48, 0x65, 0x6c, 0x6c, 0x6f},
			want:   frameHeader{fin: true, opcode: opcodePing, payloadLength: 5},
		},
		{
			name:   "masked_pong",
			reader: []byte{0x8a, 0x85, 0x37, 0xfa, 0x21, 0x3d, 0x7f, 0x9f, 0x4d, 0x51, 0x58},
			want:   frameHeader{fin: true, opcode: opcodePong, mask: true, payloadLength: 5},
		},
		{
			name:   "256b_unmasked_binary",
			reader: []byte{0x82, 0x7e, 0x01, 0x00},
			want:   frameHeader{fin: true, opcode: OpcodeBinary, payloadLength: 256},
		},
		{
			name:   "64k_unmasked_binary",
			reader: []byte{0x82, 0x7f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00},
			want:   frameHeader{fin: true, opcode: OpcodeBinary, payloadLength: 65536},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Conn{bufio: bufio.NewReadWriter(bufio.NewReader(bytes.NewReader(tt.reader)), nil)}
			got, err := c.readFrameHeader()
			if (err != nil) != tt.wantErr {
				t.Errorf("Conn.readFrameHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Conn.readFrameHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConnWriteFrame(t *testing.T) {
	c := &Conn{}
	b := new(bytes.Buffer)
	c.bufio = bufio.NewReadWriter(nil, bufio.NewWriter(b))

	payload := []byte("hello")
	origPayload := []byte("hello")
	if err := c.writeFrame(OpcodeText, payload); err != nil {
		t.Fatalf("Conn.writeFrame() error = %v", err)
	}

	want := []byte{0x81, 0x85, 0, 0, 0, 0, 'h', 'e', 'l', 'l', 'o'}

	got := b.Bytes()
	for i := range 4 {
		want[2+i] = got[2+i]
	}
	for i := range payload {
		want[6+i] ^= got[2+(i%4)]
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Conn.writeFrame() output = %v, want %v", got, want)
	}

	// Input payload must no longer be masked when the function returns.
	if !reflect.DeepEqual(payload, origPayload) {
		t.Errorf("Conn.writeFrame() input = %v, want %v", payload, origPayload)
	}
}

func TestConnWritePayloadLength(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want []byte
	}{
		{
			name: "0",
			n:    0,
			want: []byte{0x80},
		},
		{
			name: "1",
			n:    1,
			want: []byte{0x80 | 1},
		},
		{
			name: "125",
			n:    125,
			want: []byte{0x80 | 125},
		},
		{
			name: "126",
			n:    126,
			want: []byte{0xfe, 0x00, 126},
		},
		{
			name: "65535",
			n:    65535,
			want: []byte{0xfe, 0xff, 0xff},
		},
		{
			name: "65536",
			n:    65536,
			want: []byte{0xff, 0, 0, 0, 0, 0, 1, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Conn{}
			b := new(bytes.Buffer)
			c.bufio = bufio.NewReadWriter(nil, bufio.NewWriter(b))

			if err := c.writePayloadLength(tt.n); err != nil {
				t.Fatalf("Conn.writePayloadLength() error = %v", err)
			}

			_ = c.bufio.Flush()

			if !reflect.DeepEqual(b.Bytes(), tt.want) {
				t.Errorf("Conn.writePayloadLength() = %v, want %v", b.Bytes(), tt.want)
			}
		})
	}
}

func TestConnMask(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		want    []byte
	}{
		{
			name: "nil_payload",
		},
		{
			name:    "empty_payload",
			payload: []byte{},
			want:    []byte{},
		},
		{
			name:    "1_byte",
			payload: []byte("a"),
			want:    []byte{88},
		},
		{
			name:    "4_bytes",
			payload: []byte("abcd"),
			want:    []byte{88, 90, 84, 82},
		},
		{
			name:    "inverse_of_4_bytes",
			payload: []byte{88, 90, 84, 82},
			want:    []byte("abcd"),
		},
		{
			name:    "6_bytes",
			payload: []byte("abcdef"),
			want:    []byte{88, 90, 84, 82, 92, 94},
		},
		{
			name:    "8_bytes",
			payload: []byte("abcdefgh"),
			want:    []byte{88, 90, 84, 82, 92, 94, 80, 94},
		},
		{
			name:    "10_bytes",
			payload: []byte("abcdefghij"),
			want:    []byte{88, 90, 84, 82, 92, 94, 80, 94, 80, 82},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Conn{}
			copy(c.writeBuf[:4], []byte("9876"))

			c.mask(tt.payload)
			if !reflect.DeepEqual(tt.payload, tt.want) {
				t.Errorf("Conn.mask() = %v, want %v", tt.payload, tt.want)
			}
		})
	}
}
