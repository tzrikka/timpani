package slack

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/tzrikka/timpani-api/pkg/slack"
	"github.com/tzrikka/timpani/internal/listeners"
	"github.com/tzrikka/timpani/pkg/metrics"
)

// https://docs.slack.dev/reference/methods/chat.delete/
func (a *API) ChatDeleteActivity(ctx context.Context, req slack.ChatDeleteRequest) (*slack.ChatDeleteResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.ChatDeleteResponse)
	if err := a.httpPost(ctx, slack.ChatDeleteActivityName, req, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.ChatDeleteActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.ChatDeleteActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.ChatDeleteActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/chat.getPermalink/
func (a *API) ChatGetPermalinkActivity(ctx context.Context, req slack.ChatGetPermalinkRequest) (*slack.ChatGetPermalinkResponse, error) {
	query := url.Values{}
	query.Set("channel", req.Channel)
	query.Set("message_ts", req.MessageTS)

	t := time.Now().UTC()
	resp := new(slack.ChatGetPermalinkResponse)
	if err := a.httpGet(ctx, slack.ChatGetPermalinkActivityName, query, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.ChatGetPermalinkActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.ChatGetPermalinkActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.ChatGetPermalinkActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/chat.postEphemeral/
func (a *API) ChatPostEphemeralActivity(ctx context.Context, req slack.ChatPostEphemeralRequest) (*slack.ChatPostEphemeralResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.ChatPostEphemeralResponse)
	if err := a.httpPost(ctx, slack.ChatPostEphemeralActivityName, req, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.ChatPostEphemeralActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.ChatPostEphemeralActivityName, slackAPIError(resp, resp.Error))

		switch resp.Error {
		case "channel_not_found", "is_archived", "not_in_channel", "user_not_in_channel":
			return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, req.Channel, resp)
		}
		if strings.Contains(resp.Error, "invalid") {
			return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, resp)
		}
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.ChatPostEphemeralActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/chat.postMessage/
func (a *API) ChatPostMessageActivity(ctx context.Context, req slack.ChatPostMessageRequest) (*slack.ChatPostMessageResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.ChatPostMessageResponse)
	if err := a.httpPost(ctx, slack.ChatPostMessageActivityName, req, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.ChatPostMessageActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.ChatPostMessageActivityName, slackAPIError(resp, resp.Error))

		switch resp.Error {
		case "channel_not_found", "is_archived", "not_in_channel", "user_not_in_channel":
			return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, req.Channel, resp)
		}
		if strings.Contains(resp.Error, "invalid") {
			return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, resp)
		}
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.ChatPostMessageActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/chat.update/
func (a *API) ChatUpdateActivity(ctx context.Context, req slack.ChatUpdateRequest) (*slack.ChatUpdateResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.ChatUpdateResponse)
	if err := a.httpPost(ctx, slack.ChatUpdateActivityName, req, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.ChatUpdateActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.ChatUpdateActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.ChatUpdateActivityName, nil)
	return resp, nil
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
	actx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		TaskQueue:           info.TaskQueueName,
		StartToCloseTimeout: 5 * time.Second,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 5},
	})

	fut1 := workflow.ExecuteActivity(actx, slack.ChatPostMessageActivityName, slack.ChatPostMessageRequest{
		Channel: req.Channel,
		Blocks:  approvalBlocks(req, id),

		ThreadTS:       req.ThreadTS,
		ReplyBroadcast: req.ReplyBroadcast,
		Metadata:       req.Metadata,
	})

	if err := fut1.Get(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to post chat message: %w", err)
	}

	fut2 := workflow.ExecuteChildWorkflow(ctx, listeners.WaitForEventWorkflow, listeners.WaitForEventRequest{
		Signal:  "slack.events.block_actions",
		Timeout: req.Timeout,
	})

	var payload map[string]any
	if err := fut2.Get(ctx, &payload); err != nil {
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
