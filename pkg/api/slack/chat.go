package slack

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/tzrikka/timpani/internal/listeners"
)

const (
	ChatDeleteName        = "slack.chat.delete"
	ChatGetPermalinkName  = "slack.chat.getPermalink"
	ChatPostEphemeralName = "slack.chat.postEphemeral"
	ChatPostMessageName   = "slack.chat.postMessage"
	ChatUpdateName        = "slack.chat.update"

	TimpaniPostApprovalName = "slack.timpani.postApproval"
)

// https://docs.slack.dev/reference/methods/chat.delete
type ChatDeleteRequest struct {
	Channel string `json:"channel"`
	TS      string `json:"ts"`

	AsUser bool `json:"as_user,omitempty"`
}

// https://docs.slack.dev/reference/methods/chat.delete
type ChatDeleteResponse struct {
	slackResponse

	Channel string `json:"channel,omitempty"`
	TS      string `json:"ts,omitempty"`
}

// https://docs.slack.dev/reference/methods/chat.delete
func (a *API) ChatDeleteActivity(ctx context.Context, req ChatDeleteRequest) (*ChatDeleteResponse, error) {
	resp := new(ChatDeleteResponse)
	if err := a.httpPost(ctx, ChatDeleteName, req, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/chat.getPermalink
type ChatGetPermalinkRequest struct {
	Channel   string `json:"channel"`
	MessageTS string `json:"message_ts"`
}

// https://docs.slack.dev/reference/methods/chat.getPermalink
type ChatGetPermalinkResponse struct {
	slackResponse

	Channel   string `json:"channel,omitempty"`
	Permalink string `json:"permalink,omitempty"`
}

// https://docs.slack.dev/reference/methods/chat.getPermalink
func (a *API) ChatGetPermalinkActivity(ctx context.Context, req ChatGetPermalinkRequest) (*ChatGetPermalinkResponse, error) {
	query := url.Values{}
	query.Set("channel", req.Channel)
	query.Set("message_ts", req.MessageTS)

	resp := new(ChatGetPermalinkResponse)
	if err := a.httpGet(ctx, ChatGetPermalinkName, query, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/chat.postEphemeral
//
// https://docs.slack.dev/reference/methods/chat.postMessage#channels
type ChatPostEphemeralRequest struct {
	Channel string `json:"channel"`
	User    string `json:"user"`

	Attachments  []map[string]any `json:"attachments,omitempty"`
	Blocks       []map[string]any `json:"blocks,omitempty"`
	IconEmoji    string           `json:"icon_emoji,omitempty"`
	IconURL      string           `json:"icon_url,omitempty"`
	LinkNames    bool             `json:"link_names,omitempty"`
	MarkdownText string           `json:"markdown_text,omitempty"`
	Parse        string           `json:"parse,omitempty"`
	Text         string           `json:"text,omitempty"`
	ThreadTS     string           `json:"thread_ts,omitempty"`
	Username     string           `json:"username,omitempty"`
}

// https://docs.slack.dev/reference/methods/chat.postEphemeral
type ChatPostEphemeralResponse struct {
	slackResponse

	MessageTS string `json:"message_ts,omitempty"`
}

// https://docs.slack.dev/reference/methods/chat.postEphemeral
func (a *API) ChatPostEphemeralActivity(ctx context.Context, req ChatPostEphemeralRequest) (*ChatPostEphemeralResponse, error) {
	resp := new(ChatPostEphemeralResponse)
	if err := a.httpPost(ctx, ChatPostEphemeralName, req, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/chat.postMessage
type ChatPostMessageRequest struct {
	Channel string `json:"channel"`

	Attachments  []map[string]any `json:"attachments,omitempty"`
	Blocks       []map[string]any `json:"blocks,omitempty"`
	IconEmoji    string           `json:"icon_emoji,omitempty"`
	IconURL      string           `json:"icon_url,omitempty"`
	LinkNames    bool             `json:"link_names,omitempty"`
	MarkdownText string           `json:"markdown_text,omitempty"`
	Metadata     map[string]any   `json:"metadata,omitempty"`
	// Ignoring "mrkdwn" for now, because it has an unusual default value (true).
	Parse          string `json:"parse,omitempty"`
	ReplyBroadcast bool   `json:"reply_broadcast,omitempty"`
	Text           string `json:"text,omitempty"`
	ThreadTS       string `json:"thread_ts,omitempty"`
	UnfurnLinks    bool   `json:"unfurl_links,omitempty"`
	Username       string `json:"username,omitempty"`
}

// https://docs.slack.dev/reference/methods/chat.postMessage
type ChatPostMessageResponse struct {
	slackResponse

	Channel string         `json:"channel,omitempty"`
	TS      string         `json:"ts,omitempty"`
	Message map[string]any `json:"message,omitempty"`
}

// https://docs.slack.dev/reference/methods/chat.postMessage
func (a *API) ChatPostMessageActivity(ctx context.Context, req ChatPostMessageRequest) (*ChatPostMessageResponse, error) {
	resp := new(ChatPostMessageResponse)
	if err := a.httpPost(ctx, ChatPostMessageName, req, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/chat.update
//
// https://docs.slack.dev/reference/methods/chat.postMessage#channels
type ChatUpdateRequest struct {
	Channel string `json:"channel"`
	TS      string `json:"ts"`

	Attachments    []map[string]any `json:"attachments,omitempty"`
	Blocks         []map[string]any `json:"blocks,omitempty"`
	MarkdownText   string           `json:"markdown_text,omitempty"`
	Metadata       map[string]any   `json:"metadata,omitempty"`
	LinkNames      bool             `json:"link_names,omitempty"`
	Parse          string           `json:"parse,omitempty"`
	Text           string           `json:"text,omitempty"`
	ReplyBroadcast bool             `json:"reply_broadcast,omitempty"`
	FileIDs        []string         `json:"file_ids,omitempty"`
}

// https://docs.slack.dev/reference/methods/chat.update
type ChatUpdateResponse struct {
	slackResponse

	Channel string         `json:"channel,omitempty"`
	TS      string         `json:"ts,omitempty"`
	Text    string         `json:"text,omitempty"`
	Message map[string]any `json:"message,omitempty"`
}

// https://docs.slack.dev/reference/methods/chat.update
func (a *API) ChatUpdateActivity(ctx context.Context, req ChatUpdateRequest) (*ChatUpdateResponse, error) {
	resp := new(ChatUpdateResponse)
	if err := a.httpPost(ctx, ChatUpdateName, req, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

const (
	DefaultGreenButton = "Approve"
	DefaultRedButton   = "Deny"
)

// TimpaniPostApprovalRequest is very similar to [ChatPostMessageRequest].
// If button labels are not specified here, their default values are
// [DefaultGreenButton] and [DefaultRedButton].
type TimpaniPostApprovalRequest struct {
	Channel string `json:"channel"`

	Header  string `json:"header"`
	Message string `json:"message"`

	GreenButton string `json:"green_button,omitempty"`
	RedButton   string `json:"red_button,omitempty"`

	Metadata       map[string]any `json:"metadata,omitempty"`
	ReplyBroadcast bool           `json:"reply_broadcast,omitempty"`
	ThreadTS       string         `json:"thread_ts,omitempty"`

	Timeout string `json:"timeout,omitempty"`
}

type TimpaniPostApprovalResponse struct {
	slackResponse

	InteractionEvent map[string]any `json:"interaction_event,omitempty"`
}

// TimpaniPostApprovalWorkflow is a convenience wrapper over
// [ChatPostMessageActivity]. It sends an interactive message to a
// user/group/channel with a short header, a markdown message, and
// 2 buttons. It then waits for (and returns) the user selection.
//
// For message formatting tips, see
// https://docs.slack.dev/messaging/formatting-message-text.
func (a *API) TimpaniPostApprovalWorkflow(ctx workflow.Context, req TimpaniPostApprovalRequest) (*TimpaniPostApprovalResponse, error) {
	info := workflow.GetInfo(ctx)
	// See the usage of action IDs in [slackEventsWorkflow].
	id := base64.RawURLEncoding.EncodeToString([]byte(info.WorkflowExecution.ID))

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		TaskQueue:           info.TaskQueueName,
		StartToCloseTimeout: 3 * time.Second,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 5},
	})

	ac := workflow.ExecuteActivity(ctx, ChatPostMessageName, ChatPostMessageRequest{
		Channel:        req.Channel,
		Blocks:         approvalBlocks(req, id),
		Metadata:       req.Metadata,
		ReplyBroadcast: req.ReplyBroadcast,
		ThreadTS:       req.ThreadTS,
	})

	resp1 := ChatPostMessageResponse{}
	if err := ac.Get(ctx, &resp1); err != nil {
		return nil, fmt.Errorf("failed to post chat message: %w", err)
	}

	wf := workflow.ExecuteChildWorkflow(ctx, listeners.WaitForEventWorkflow, listeners.WaitForEventRequest{
		Signal:  "slack.events.block_actions",
		Timeout: req.Timeout,
	})

	var payload map[string]any
	if err := wf.Get(ctx, &payload); err != nil {
		return nil, fmt.Errorf("failed to wait for events: %w", err)
	}

	return &TimpaniPostApprovalResponse{InteractionEvent: payload}, nil
}

// approvalBlocks is based on https://docs.slack.dev/block-kit.
func approvalBlocks(req TimpaniPostApprovalRequest, id string) []map[string]any {
	greenButton := req.GreenButton
	if greenButton == "" {
		greenButton = DefaultGreenButton
	}

	redButton := req.RedButton
	if redButton == "" {
		redButton = DefaultRedButton
	}

	return []map[string]any{
		{
			"type": "header",
			"text": map[string]any{
				"type":  "plain_text",
				"text":  req.Header,
				"emoji": true,
			},
		},
		{
			"type": "divider",
		},
		{
			"type": "section",
			"text": map[string]string{
				"type": "mrkdwn",
				"text": req.Message,
			},
		},
		{
			"type": "divider",
		},
		{
			"type": "actions",
			"elements": []map[string]any{
				{
					"type":  "button",
					"style": "primary",
					"text": map[string]any{
						"type":  "plain_text",
						"text":  greenButton,
						"emoji": true,
					},
					"value":     "approve",
					"action_id": "id1_" + id,
				},
				{
					"type":  "button",
					"style": "danger",
					"text": map[string]any{
						"type":  "plain_text",
						"text":  redButton,
						"emoji": true,
					},
					"value":     "deny",
					"action_id": "id2_" + id,
				},
			},
		},
	}
}
