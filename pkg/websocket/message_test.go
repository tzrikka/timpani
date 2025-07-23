package websocket

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"io"
	"testing"

	"github.com/rs/zerolog"
)

type benchmark struct {
	name      string
	msgLen    int
	bufLen    int
	frameLens []int
	frames    int
}

func BenchmarkReadMessage(b *testing.B) {
	benchmarks := []benchmark{
		{
			name:      "one_125b_frame",
			msgLen:    125,
			bufLen:    2 + 125,
			frameLens: []int{125},
			frames:    1,
		},
		{
			name:      "one_126b_frame",
			msgLen:    126,
			bufLen:    2 + 2 + 126,
			frameLens: []int{len16bits, 126},
			frames:    1,
		},
		{
			name:      "one_250b_frame",
			msgLen:    250,
			bufLen:    2 + 2 + 250,
			frameLens: []int{len16bits, 250},
			frames:    1,
		},
		{
			name:      "one_32k_frame",
			msgLen:    32768,
			bufLen:    2 + 2 + 32768,
			frameLens: []int{len16bits, 32768},
			frames:    1,
		},
		{
			name:      "one_64k-1_frame",
			msgLen:    65535,
			bufLen:    2 + 2 + 65535,
			frameLens: []int{len16bits, 65535},
			frames:    1,
		},
		{
			name:      "one_64k_frame",
			msgLen:    65536,
			bufLen:    2 + 8 + 65536,
			frameLens: []int{len64bits, 65536},
			frames:    1,
		},
		{
			name:      "one_128k_frame",
			msgLen:    131072,
			bufLen:    2 + 8 + 131072,
			frameLens: []int{len64bits, 131072},
			frames:    1,
		},
		{
			name:      "two_125b_frames",
			msgLen:    125 * 2,
			bufLen:    (2 + 125) * 2,
			frameLens: []int{125},
			frames:    2,
		},
		{
			name:      "two_32k_frames",
			msgLen:    32768 * 2,
			bufLen:    (2 + 2 + 32768) * 2,
			frameLens: []int{len16bits, 32768},
			frames:    2,
		},
		{
			name:      "two_64k_frames",
			msgLen:    65536 * 2,
			bufLen:    (2 + 8 + 65536) * 2,
			frameLens: []int{len64bits, 65536},
			frames:    2,
		},
	}

	l := zerolog.Nop()
	c := &Conn{logger: &l}

	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			f := constructBenchmarkFrame(b, bb)
			for b.Loop() {
				c.bufio = bufio.NewReadWriter(bufio.NewReader(bytes.NewReader(f)), nil)
				msg := c.readMessage()
				if n := len(msg.Data); n != bb.msgLen {
					b.Fatalf("len(msg): got %d, want %d", n, bb.msgLen)
				}
			}
		})
	}
}

func constructBenchmarkFrame(b *testing.B, bb benchmark) []byte {
	b.Helper()

	frame := make([]byte, bb.bufLen)
	i := 0
	if bb.frames == 1 {
		frame[i] = 0x82 // Binary data with FIN.
	} else if i == 0 {
		frame[i] = 0x02 // Binary data without FIN.
	}
	frame[i+1] = byte(bb.frameLens[0])
	i += 2

	switch bb.frameLens[0] {
	case len16bits:
		binary.BigEndian.PutUint16(frame[i:i+2], uint16(bb.frameLens[1])) //gosec:disable G115 -- value checked before cast
		_, _ = io.ReadFull(rand.Reader, frame[i+2:])
		i += 2 + bb.frameLens[1]
	case len64bits:
		binary.BigEndian.PutUint64(frame[i:i+8], uint64(bb.frameLens[1])) //gosec:disable G115 -- value checked before cast
		_, _ = io.ReadFull(rand.Reader, frame[i+8:])
		i += 8 + bb.frameLens[1]
	default: // Up to 125 bytes.
		_, _ = io.ReadFull(rand.Reader, frame[i:])
		i += bb.frameLens[0]
	}

	if bb.frames == 1 {
		return frame
	}

	frame[i] = 0x80 // Continuation with FIN.
	frame[i+1] = byte(bb.frameLens[0])
	i += 2

	switch bb.frameLens[0] {
	case len16bits:
		binary.BigEndian.PutUint16(frame[i:i+2], uint16(bb.frameLens[1])) //gosec:disable G115 -- value checked before cast
	case len64bits:
		binary.BigEndian.PutUint64(frame[i:i+8], uint64(bb.frameLens[1])) //gosec:disable G115 -- value checked before cast
	}

	return frame
}
