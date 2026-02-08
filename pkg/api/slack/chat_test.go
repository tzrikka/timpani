package slack

import (
	"testing"

	"github.com/tzrikka/timpani-api/pkg/slack"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name      string
		s         string
		maxLength int
		want      string
	}{
		{
			name:      "ascii",
			s:         "This is a long string to truncate",
			maxLength: 20,
			want:      "This is (truncated)",
		},
		{
			name:      "multi_byte_characters",
			s:         "こんにちは世界 means 'Hello world' in Japanese",
			maxLength: 20,
			want:      "こん (truncated)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncate(tt.s, tt.maxLength); got != tt.want {
				t.Errorf("truncate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestApprovalBlocks(t *testing.T) {
	blocks := approvalBlocks(slack.TimpaniPostApprovalRequest{RedButton: "red"}, "id")

	elems, ok := blocks[4]["elements"]
	if !ok {
		t.Fatal(`approvalBlocks() last block doesn't contain "elements" key`)
	}

	es, ok := elems.([]map[string]any)
	if !ok {
		t.Fatalf("approvalBlocks() last block's elements type: got %T, want []map[string]any", elems)
	}

	want := DefaultGreenButton
	if got := es[0]["text"].(map[string]any)["text"]; got != want { //nolint:errcheck // Type conversion always succeeds.
		t.Errorf("approvalBlocks() green button label = %q, want %q", got, want)
	}

	want = "red"
	if got := es[1]["text"].(map[string]any)["text"]; got != want { //nolint:errcheck // Type conversion always succeeds.
		t.Errorf("approvalBlocks() red button label = %q, want %q", got, want)
	}

	if es[0]["action_id"] == es[1]["action_id"] {
		t.Error("approvalBlocks() button action IDs must be unique")
	}
}
