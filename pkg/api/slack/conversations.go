package slack

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.temporal.io/sdk/temporal"

	"github.com/tzrikka/timpani-api/pkg/slack"
	"github.com/tzrikka/timpani/pkg/metrics"
)

// https://docs.slack.dev/reference/methods/conversations.archive/
func (a *API) ConversationsArchiveActivity(ctx context.Context, req slack.ConversationsArchiveRequest) (*slack.ConversationsArchiveResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.ConversationsArchiveResponse)
	if err := a.httpPost(ctx, slack.ConversationsArchiveActivityName, req, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.ConversationsArchiveActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.ConversationsArchiveActivityName, slackAPIError(resp, resp.Error))

		if resp.Error == "already_archived" {
			return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, req.Channel)
		}
		if strings.Contains(resp.Error, "invalid") {
			return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, req.Channel, resp)
		}
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.ConversationsArchiveActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.close/
func (a *API) ConversationsCloseActivity(ctx context.Context, req slack.ConversationsCloseRequest) (*slack.ConversationsCloseResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.ConversationsCloseResponse)
	if err := a.httpPost(ctx, slack.ConversationsCloseActivityName, req, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.ConversationsCloseActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.ConversationsCloseActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.ConversationsCloseActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.create/
func (a *API) ConversationsCreateActivity(ctx context.Context, req slack.ConversationsCreateRequest) (*slack.ConversationsCreateResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.ConversationsCreateResponse)
	if err := a.httpPost(ctx, slack.ConversationsCreateActivityName, req, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.ConversationsCreateActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.ConversationsCreateActivityName, slackAPIError(resp, resp.Error))

		if resp.Error == "name_taken" {
			return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, req.Name)
		}
		if strings.Contains(resp.Error, "invalid") {
			return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, resp)
		}
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.ConversationsCreateActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.history/
func (a *API) ConversationsHistoryActivity(ctx context.Context, req slack.ConversationsHistoryRequest) (*slack.ConversationsHistoryResponse, error) {
	query := url.Values{}
	query.Set("channel", req.Channel)
	if req.IncludeAllMetadata {
		query.Set("include_all_metadata", "true")
	}
	if req.Inclusive {
		query.Set("inclusive", "true")
	}
	if req.Latest != "" {
		query.Set("latest", req.Latest)
	}
	if req.Oldest != "" {
		query.Set("oldest", req.Oldest)
	}
	if req.Limit != 0 {
		query.Set("limit", strconv.Itoa(req.Limit))
	}
	if req.Cursor != "" {
		query.Set("cursor", req.Cursor)
	}

	t := time.Now().UTC()
	resp := new(slack.ConversationsHistoryResponse)
	if err := a.httpGet(ctx, slack.ConversationsHistoryActivityName, query, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.ConversationsHistoryActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.ConversationsHistoryActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.ConversationsHistoryActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.info/
func (a *API) ConversationsInfoActivity(ctx context.Context, req slack.ConversationsInfoRequest) (*slack.ConversationsInfoResponse, error) {
	query := url.Values{}
	query.Set("channel", req.Channel)
	if req.IncludeLocale {
		query.Set("include_locale", "true")
	}
	if req.IncludeNumMembers {
		query.Set("include_num_members", "true")
	}

	t := time.Now().UTC()
	resp := new(slack.ConversationsInfoResponse)
	if err := a.httpGet(ctx, slack.ConversationsInfoActivityName, query, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.ConversationsInfoActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.ConversationsInfoActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.ConversationsInfoActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.invite/
func (a *API) ConversationsInviteActivity(ctx context.Context, req slack.ConversationsInviteRequest) (*slack.ConversationsInviteResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.ConversationsInviteResponse)
	if err := a.httpPost(ctx, slack.ConversationsInviteActivityName, req, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.ConversationsInviteActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.ConversationsInviteActivityName, slackAPIError(resp, resp.Error))

		if resp.Error == "already_in_channel" {
			return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, req, resp)
		}
		if strings.Contains(resp.Error, "invalid") {
			return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, req, resp)
		}
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.ConversationsInviteActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.join/
func (a *API) ConversationsJoinActivity(ctx context.Context, req slack.ConversationsJoinRequest) (*slack.ConversationsJoinResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.ConversationsJoinResponse)
	if err := a.httpPost(ctx, slack.ConversationsJoinActivityName, req, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.ConversationsJoinActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.ConversationsJoinActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.ConversationsJoinActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.kick/
func (a *API) ConversationsKickActivity(ctx context.Context, req slack.ConversationsKickRequest) (*slack.ConversationsKickResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.ConversationsKickResponse)
	if err := a.httpPost(ctx, slack.ConversationsKickActivityName, req, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.ConversationsKickActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.ConversationsKickActivityName, slackAPIError(resp, resp.Error))

		switch resp.Error {
		case "channel_not_found", "not_in_channel":
			return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, req.Channel, resp)
		}
		if strings.Contains(resp.Error, "invalid") {
			return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, req, resp)
		}
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.ConversationsKickActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.leave/
func (a *API) ConversationsLeaveActivity(ctx context.Context, req slack.ConversationsLeaveRequest) (*slack.ConversationsLeaveResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.ConversationsLeaveResponse)
	if err := a.httpPost(ctx, slack.ConversationsLeaveActivityName, req, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.ConversationsLeaveActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.ConversationsLeaveActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.ConversationsLeaveActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.list/
func (a *API) ConversationsListActivity(ctx context.Context, req slack.ConversationsListRequest) (*slack.ConversationsListResponse, error) {
	query := url.Values{}
	if req.Types != "" {
		query.Set("types", req.Types)
	}
	if req.ExcludeArchived {
		query.Set("exclude_archived", "true")
	}
	if req.Limit != 0 {
		query.Set("limit", strconv.Itoa(req.Limit))
	}
	if req.Cursor != "" {
		query.Set("cursor", req.Cursor)
	}
	if req.TeamID != "" {
		query.Set("team_id", req.TeamID)
	}

	t := time.Now().UTC()
	resp := new(slack.ConversationsListResponse)
	if err := a.httpGet(ctx, slack.ConversationsListActivityName, query, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.ConversationsListActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.ConversationsListActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.ConversationsListActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.members/
func (a *API) ConversationsMembersActivity(ctx context.Context, req slack.ConversationsMembersRequest) (*slack.ConversationsMembersResponse, error) {
	query := url.Values{}
	query.Set("channel", req.Channel)
	if req.Cursor != "" {
		query.Set("cursor", req.Cursor)
	}
	if req.Limit != 0 {
		query.Set("limit", strconv.Itoa(req.Limit))
	}

	t := time.Now().UTC()
	resp := new(slack.ConversationsMembersResponse)
	if err := a.httpGet(ctx, slack.ConversationsMembersActivityName, query, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.ConversationsMembersActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.ConversationsMembersActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.ConversationsMembersActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.open/
func (a *API) ConversationsOpenActivity(ctx context.Context, req slack.ConversationsOpenRequest) (*slack.ConversationsOpenResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.ConversationsOpenResponse)
	if err := a.httpPost(ctx, slack.ConversationsOpenActivityName, req, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.ConversationsOpenActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.ConversationsOpenActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.ConversationsOpenActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.rename/
func (a *API) ConversationsRenameActivity(ctx context.Context, req slack.ConversationsRenameRequest) (*slack.ConversationsRenameResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.ConversationsRenameResponse)
	if err := a.httpPost(ctx, slack.ConversationsRenameActivityName, req, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.ConversationsRenameActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.ConversationsRenameActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.ConversationsRenameActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.replies/
func (a *API) ConversationsRepliesActivity(ctx context.Context, req slack.ConversationsRepliesRequest) (*slack.ConversationsRepliesResponse, error) {
	query := url.Values{}
	query.Set("channel", req.Channel)
	query.Set("ts", req.TS)
	if req.IncludeAllMetadata {
		query.Set("include_all_metadata", "true")
	}
	if req.Inclusive {
		query.Set("inclusive", "true")
	}
	if req.Latest != "" {
		query.Set("latest", req.Latest)
	}
	if req.Oldest != "" {
		query.Set("oldest", req.Oldest)
	}
	if req.Limit != 0 {
		query.Set("limit", strconv.Itoa(req.Limit))
	}
	if req.Cursor != "" {
		query.Set("cursor", req.Cursor)
	}

	t := time.Now().UTC()
	resp := new(slack.ConversationsRepliesResponse)
	if err := a.httpGet(ctx, slack.ConversationsRepliesActivityName, query, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.ConversationsRepliesActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.ConversationsRepliesActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.ConversationsRepliesActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.setPurpose/
func (a *API) ConversationsSetPurposeActivity(
	ctx context.Context,
	req slack.ConversationsSetPurposeRequest,
) (*slack.ConversationsSetPurposeResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.ConversationsSetPurposeResponse)
	if err := a.httpPost(ctx, slack.ConversationsSetPurposeActivityName, req, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.ConversationsSetPurposeActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.ConversationsSetPurposeActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.ConversationsSetPurposeActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.setTopic/
func (a *API) ConversationsSetTopicActivity(
	ctx context.Context,
	req slack.ConversationsSetTopicRequest,
) (*slack.ConversationsSetTopicResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.ConversationsSetTopicResponse)
	if err := a.httpPost(ctx, slack.ConversationsSetTopicActivityName, req, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.ConversationsSetTopicActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.ConversationsSetTopicActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.ConversationsSetTopicActivityName, nil)
	return resp, nil
}
