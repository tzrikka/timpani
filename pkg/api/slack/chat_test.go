package slack

import (
	"testing"

	"github.com/tzrikka/timpani-api/pkg/slack"
)

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
	if got := es[0]["text"].(map[string]any)["text"]; got != want {
		t.Errorf("approvalBlocks() green button label = %q, want %q", got, want)
	}

	want = "red"
	if got := es[1]["text"].(map[string]any)["text"]; got != want {
		t.Errorf("approvalBlocks() green button label = %q, want %q", got, want)
	}

	if es[0]["action_id"] == es[1]["action_id"] {
		t.Error("approvalBlocks() button action IDs must be unique")
	}
}
