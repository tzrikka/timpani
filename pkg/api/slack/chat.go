package slack

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/tzrikka/timpani-api/pkg/slack"
	"github.com/tzrikka/timpani/internal/listeners"
)

const (
	// MarkdownTextMaxLength is based on:
	//   - https://docs.slack.dev/reference/methods/chat.postEphemeral/#arguments
	//   - https://docs.slack.dev/reference/methods/chat.postMessage/#arguments
	//   - https://docs.slack.dev/reference/methods/chat.update/#arguments
	MarkdownTextMaxLength = 12000

	// UpdateTextMaxLength is based on:
	// https://docs.slack.dev/reference/methods/chat.update/#errors (msg_too_long).
	UpdateTextMaxLength = 4000
)

// ChatDeleteActivity is based on:
// https://docs.slack.dev/reference/methods/chat.delete/
func (a *API) ChatDeleteActivity(ctx context.Context, req slack.ChatDeleteRequest) (*slack.ChatDeleteResponse, error) {
	resp := new(slack.ChatDeleteResponse)
	if err := a.httpPost(ctx, slack.ChatDeleteActivityName, req, resp); err != nil {
		return nil, err
	}

	if resp.Error == "cant_delete_message" {
		return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, req.Channel, req.TS, resp)
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// ChatGetPermalinkActivity is based on:
// https://docs.slack.dev/reference/methods/chat.getPermalink/
func (a *API) ChatGetPermalinkActivity(ctx context.Context, req slack.ChatGetPermalinkRequest) (*slack.ChatGetPermalinkResponse, error) {
	query := url.Values{}
	query.Set("channel", req.Channel)
	query.Set("message_ts", req.MessageTS)

	resp := new(slack.ChatGetPermalinkResponse)
	if err := a.httpGet(ctx, slack.ChatGetPermalinkActivityName, query, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// ChatPostEphemeralActivity is based on:
// https://docs.slack.dev/reference/methods/chat.postEphemeral/
func (a *API) ChatPostEphemeralActivity(ctx context.Context, req slack.ChatPostEphemeralRequest) (*slack.ChatPostEphemeralResponse, error) {
	if l := len(req.MarkdownText); l > MarkdownTextMaxLength {
		activity.GetLogger(ctx).Warn("truncating Slack message markdown",
			slog.Int("original_length", l), slog.Int("new_length", MarkdownTextMaxLength))
		req.MarkdownText = truncate(req.MarkdownText, MarkdownTextMaxLength)
	}

	resp := new(slack.ChatPostEphemeralResponse)
	if err := a.httpPost(ctx, slack.ChatPostEphemeralActivityName, req, resp); err != nil {
		return nil, err
	}

	switch resp.Error {
	case "channel_not_found", "is_archived", "not_in_channel", "user_not_in_channel":
		return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, req.Channel, req.User)
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// ChatPostMessageActivity is based on:
// https://docs.slack.dev/reference/methods/chat.postMessage/
func (a *API) ChatPostMessageActivity(ctx context.Context, req slack.ChatPostMessageRequest) (*slack.ChatPostMessageResponse, error) {
	if l := len(req.MarkdownText); l > MarkdownTextMaxLength {
		activity.GetLogger(ctx).Warn("truncating Slack message markdown",
			slog.Int("original_length", l), slog.Int("new_length", MarkdownTextMaxLength))
		req.MarkdownText = truncate(req.MarkdownText, MarkdownTextMaxLength)
	}

	resp := new(slack.ChatPostMessageResponse)
	if err := a.httpPost(ctx, slack.ChatPostMessageActivityName, req, resp); err != nil {
		return nil, err
	}

	switch resp.Error {
	case "channel_not_found", "is_archived", "not_in_channel", "user_not_in_channel":
		return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, req.Channel)
	case "msg_too_long":
		req.Text = strconv.Itoa(len(req.Text))
		req.MarkdownText = strconv.Itoa(len(req.MarkdownText))
		return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, req)
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// ChatUpdateActivity is based on:
// https://docs.slack.dev/reference/methods/chat.update/
func (a *API) ChatUpdateActivity(ctx context.Context, req slack.ChatUpdateRequest) (*slack.ChatUpdateResponse, error) {
	if l := len(req.MarkdownText); l > MarkdownTextMaxLength {
		activity.GetLogger(ctx).Warn("truncating Slack message markdown",
			slog.Int("original_length", l), slog.Int("new_length", MarkdownTextMaxLength))
		req.MarkdownText = truncate(req.MarkdownText, MarkdownTextMaxLength)
	}
	if l := len(req.Text); l > UpdateTextMaxLength {
		activity.GetLogger(ctx).Warn("truncating Slack message text",
			slog.Int("original_length", l), slog.Int("new_length", UpdateTextMaxLength))
		req.Text = truncate(req.Text, UpdateTextMaxLength)
	}

	resp := new(slack.ChatUpdateResponse)
	if err := a.httpPost(ctx, slack.ChatUpdateActivityName, req, resp); err != nil {
		return nil, err
	}

	switch resp.Error {
	case "cant_update_message":
		return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, req.Channel, req.TS, resp)
	case "msg_too_long":
		req.Text = strconv.Itoa(len(req.Text))
		req.MarkdownText = strconv.Itoa(len(req.MarkdownText))
		return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, req, resp)
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// truncate truncates the input string to the specified maximum length.
// It ensures that we do not truncate in the middle of a multi-byte character.
func truncate(s string, maxLength int) string {
	maxLength = maxLength - 12 // For the " (truncated)" suffix.
	l := len(s) - maxLength
	r := []rune(s)

	for l > 0 {
		r = r[:len(r)-int(math.Max(1, float64(l/4)))]
		l = len(string(r)) - maxLength
	}

	return strings.TrimSpace(string(r)) + " (truncated)"
}

// TimpaniPostApprovalWorkflow is a convenience wrapper over
// [ChatPostMessageActivity]. It sends an interactive message to a
// user/group/channel with a short header, a markdown message, and
// 2 buttons. It then waits for (and returns) the user selection.
//
// For message formatting tips, see
// https://docs.slack.dev/messaging/formatting-message-text.
func (a *API) TimpaniPostApprovalWorkflow(ctx workflow.Context, req slack.TimpaniPostApprovalRequest) (*slack.TimpaniPostApprovalResponse, error) {
	info := workflow.GetInfo(ctx)
	id := base64.RawURLEncoding.EncodeToString([]byte(info.WorkflowExecution.ID))
	txCallCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		TaskQueue:           info.TaskQueueName,
		StartToCloseTimeout: 5 * time.Second,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 5},
	})
	txCallFut := workflow.ExecuteActivity(txCallCtx, slack.ChatPostMessageActivityName, slack.ChatPostMessageRequest{
		Channel:        req.Channel,
		Blocks:         approvalBlocks(req, id),
		ThreadTS:       req.ThreadTS,
		ReplyBroadcast: req.ReplyBroadcast,
		Metadata:       req.Metadata,
	})

	if err := txCallFut.Get(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to post chat message: %w", err)
	}

	// https://docs.temporal.io/develop/go/observability#visibility
	signal := "slack.events.block_actions"
	attr := temporal.NewSearchAttributeKeyKeywordList("WaitingForSignals").ValueSet([]string{signal})
	opts := workflow.ChildWorkflowOptions{TypedSearchAttributes: temporal.NewSearchAttributes(attr)}

	rxEventCtx := workflow.WithChildOptions(ctx, opts)
	rxEventReq := listeners.WaitForEventRequest{Signal: signal, Timeout: req.Timeout}
	rxEventFut := workflow.ExecuteChildWorkflow(rxEventCtx, listeners.WaitForEventWorkflow, rxEventReq)

	var payload map[string]any
	if err := rxEventFut.Get(ctx, &payload); err != nil {
		return nil, fmt.Errorf("failed to wait for events: %w", err)
	}

	return &slack.TimpaniPostApprovalResponse{InteractionEvent: payload}, nil
}

const (
	DefaultGreenButton = "Approve"
	DefaultRedButton   = "Deny"
)

// approvalBlocks is based on https://docs.slack.dev/block-kit.
func approvalBlocks(req slack.TimpaniPostApprovalRequest, id string) []map[string]any {
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
