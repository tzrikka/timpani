package slack

import (
	"net/url"
	"reflect"
	"testing"
)

func TestWebFormToMap(t *testing.T) {
	tests := []struct {
		name string
		vs   url.Values
		want map[string]any
	}{
		{
			name: "empty",
			want: map[string]any{},
		},
		{
			name: "one_key_value",
			vs: url.Values{
				"key": []string{"value"},
			},
			want: map[string]any{
				"key": "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := webFormToMap(tt.vs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("webFormToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
