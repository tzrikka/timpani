package temporal

import (
	"log/slog"
	"testing"
)

func TestSanitizeSignalName(t *testing.T) {
	tests := []struct {
		name   string
		signal string
		want   string
	}{
		{
			name:   "empty",
			signal: "",
			want:   "",
		},
		{
			name:   "valid",
			signal: "aaa.bbb_ccc",
			want:   "aaa.bbb_ccc",
		},
		{
			name:   "invalid",
			signal: "invalid-name' OR 1=1; --",
			want:   "invalid_name__OR_1_1____",
		},
		{
			name:   "long",
			signal: "name_with_very_long_length_exceeding_one_hundred_characters_to_test_the_sanitization_functionality_which_should_truncate_it",
			want:   "name_with_very_long_length_exceeding_one_hundred_characters_to_test_the_sanitization_functionality_w",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeSignalName(slog.Default(), tt.signal); got != tt.want {
				t.Errorf("sanitizeSignalName() = %q, want %q", got, tt.want)
			}
		})
	}
}
