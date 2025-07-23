package temporal

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestLogMessageWithAttributes(t *testing.T) {
	tests := []struct {
		name  string
		level func() *zerolog.Event
		kv    any
	}{
		{
			name:  "debug",
			level: log.Logger.Debug,
			kv:    nil,
		},
		{
			name:  "info",
			level: log.Logger.Info,
			kv:    []string{"k1", "v1"},
		},
		{
			name:  "warn",
			level: log.Logger.Warn,
			kv:    []string{"k1", "v1", "k2", "v2"},
		},
		{
			name:  "error",
			level: log.Logger.Error,
			kv:    []string{"k1", "v1", "k2", "v2", "k3", "v3"},
		},
		{
			name:  "not_string_slice",
			level: log.Logger.Error,
			kv:    0,
		},
		{
			name:  "odd_slice_length",
			level: log.Logger.Error,
			kv:    []string{"k1", "v1", "k2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			var keyvals []any
			if tt.kv != nil {
				keyvals = []any{tt.kv}
			}
			logMessageWithAttributes(tt.level(), "msg", keyvals...)
		})
	}
}
