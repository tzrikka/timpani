package websocket

import (
	"testing"
)

func TestValidUTF8(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "ascii",
			s:    "This is an ASCII string without multi-byte characters",
			want: "This is an ASCII string without multi-byte characters",
		},
		{
			name: "valid_multi_bytes",
			s:    "こんにちは世界", //nolint:gosmopolitan // Test string.
			want: "こんにちは世界", //nolint:gosmopolitan // Test string.
		},
		{
			name: "invalid_multi_bytes",
			s:    "こんにちは世界"[:len("こんにちは世界")-1], //nolint:gosmopolitan // Test string.
			want: "こんにちは世",                     //nolint:gosmopolitan // Test string.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validUTF8(tt.s); got != tt.want {
				t.Errorf("validUTF8() = %q, want %q", got, tt.want)
			}
		})
	}
}
