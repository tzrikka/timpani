package slack

import (
	"testing"
)

func TestRandomInt(t *testing.T) {
	tests := []struct {
		name         string
		maxValue     int64
		wantLessThan int
	}{
		{
			name:         "max_1",
			maxValue:     1,
			wantLessThan: 1,
		},
		{
			name:         "max_10",
			maxValue:     10,
			wantLessThan: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := randomInt(tt.maxValue)
			if got < 0 {
				t.Errorf("randomInt() = %v, want >= 0", got)
			}
			if got >= tt.wantLessThan {
				t.Errorf("randomInt() = %v, want < %v", got, tt.wantLessThan)
			}
		})
	}
}
