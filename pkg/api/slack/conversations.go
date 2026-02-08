package slack

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	"go.temporal.io/sdk/temporal"

	"github.com/tzrikka/timpani-api/pkg/slack"
)

// ConversationsArchiveActivity is based on:
// https://docs.slack.dev/reference/methods/conversations.archive/
func (a *API) ConversationsArchiveActivity(ctx context.Context, req slack.ConversationsArchiveRequest) (*slack.ConversationsArchiveResponse, error) {
	resp := new(slack.ConversationsArchiveResponse)
	if err := a.httpPost(ctx, slack.ConversationsArchiveActivityName, req, resp); err != nil {
		return nil, err
	}

	if resp.Error == "already_archived" {
		return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, req.Channel)
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// ConversationsCloseActivity is based on:
// https://docs.slack.dev/reference/methods/conversations.close/
func (a *API) ConversationsCloseActivity(ctx context.Context, req slack.ConversationsCloseRequest) (*slack.ConversationsCloseResponse, error) {
	resp := new(slack.ConversationsCloseResponse)
	if err := a.httpPost(ctx, slack.ConversationsCloseActivityName, req, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// ConversationsCreateActivity is based on:
// https://docs.slack.dev/reference/methods/conversations.create/
func (a *API) ConversationsCreateActivity(ctx context.Context, req slack.ConversationsCreateRequest) (*slack.ConversationsCreateResponse, error) {
	resp := new(slack.ConversationsCreateResponse)
	if err := a.httpPost(ctx, slack.ConversationsCreateActivityName, req, resp); err != nil {
		return nil, err
	}

	if resp.Error == "name_taken" {
		return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, req.Name)
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// ConversationsHistoryActivity is based on:
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

	resp := new(slack.ConversationsHistoryResponse)
	if err := a.httpGet(ctx, slack.ConversationsHistoryActivityName, query, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// ConversationsInfoActivity is based on:
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

	resp := new(slack.ConversationsInfoResponse)
	if err := a.httpGet(ctx, slack.ConversationsInfoActivityName, query, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// ConversationsInviteActivity is based on:
// https://docs.slack.dev/reference/methods/conversations.invite/
func (a *API) ConversationsInviteActivity(ctx context.Context, req slack.ConversationsInviteRequest) (*slack.ConversationsInviteResponse, error) {
	resp := new(slack.ConversationsInviteResponse)
	if err := a.httpPost(ctx, slack.ConversationsInviteActivityName, req, resp); err != nil {
		return nil, err
	}

	if resp.Error == "already_in_channel" {
		return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, req, resp)
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// ConversationsJoinActivity is based on:
// https://docs.slack.dev/reference/methods/conversations.join/
func (a *API) ConversationsJoinActivity(ctx context.Context, req slack.ConversationsJoinRequest) (*slack.ConversationsJoinResponse, error) {
	resp := new(slack.ConversationsJoinResponse)
	if err := a.httpPost(ctx, slack.ConversationsJoinActivityName, req, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// ConversationsKickActivity is based on:
// https://docs.slack.dev/reference/methods/conversations.kick/
func (a *API) ConversationsKickActivity(ctx context.Context, req slack.ConversationsKickRequest) (*slack.ConversationsKickResponse, error) {
	resp := new(slack.ConversationsKickResponse)
	if err := a.httpPost(ctx, slack.ConversationsKickActivityName, req, resp); err != nil {
		return nil, err
	}

	switch resp.Error {
	case "channel_not_found", "not_in_channel":
		return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, req.Channel, req.User)
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// ConversationsLeaveActivity is based on:
// https://docs.slack.dev/reference/methods/conversations.leave/
func (a *API) ConversationsLeaveActivity(ctx context.Context, req slack.ConversationsLeaveRequest) (*slack.ConversationsLeaveResponse, error) {
	resp := new(slack.ConversationsLeaveResponse)
	if err := a.httpPost(ctx, slack.ConversationsLeaveActivityName, req, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// ConversationsListActivity is based on:
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

	resp := new(slack.ConversationsListResponse)
	if err := a.httpGet(ctx, slack.ConversationsListActivityName, query, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// ConversationsMembersActivity is based on:
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

	resp := new(slack.ConversationsMembersResponse)
	if err := a.httpGet(ctx, slack.ConversationsMembersActivityName, query, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// ConversationsOpenActivity is based on:
// https://docs.slack.dev/reference/methods/conversations.open/
func (a *API) ConversationsOpenActivity(ctx context.Context, req slack.ConversationsOpenRequest) (*slack.ConversationsOpenResponse, error) {
	resp := new(slack.ConversationsOpenResponse)
	if err := a.httpPost(ctx, slack.ConversationsOpenActivityName, req, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// ConversationsRenameActivity is based on:
// https://docs.slack.dev/reference/methods/conversations.rename/
func (a *API) ConversationsRenameActivity(ctx context.Context, req slack.ConversationsRenameRequest) (*slack.ConversationsRenameResponse, error) {
	resp := new(slack.ConversationsRenameResponse)
	if err := a.httpPost(ctx, slack.ConversationsRenameActivityName, req, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// ConversationsRepliesActivity is based on:
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

	resp := new(slack.ConversationsRepliesResponse)
	if err := a.httpGet(ctx, slack.ConversationsRepliesActivityName, query, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// ConversationsSetPurposeActivity is based on:
// https://docs.slack.dev/reference/methods/conversations.setPurpose/
func (a *API) ConversationsSetPurposeActivity(
	ctx context.Context,
	req slack.ConversationsSetPurposeRequest,
) (*slack.ConversationsSetPurposeResponse, error) {
	resp := new(slack.ConversationsSetPurposeResponse)
	if err := a.httpPost(ctx, slack.ConversationsSetPurposeActivityName, req, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// ConversationsSetTopicActivity is based on:
// https://docs.slack.dev/reference/methods/conversations.setTopic/
func (a *API) ConversationsSetTopicActivity(
	ctx context.Context,
	req slack.ConversationsSetTopicRequest,
) (*slack.ConversationsSetTopicResponse, error) {
	resp := new(slack.ConversationsSetTopicResponse)
	if err := a.httpPost(ctx, slack.ConversationsSetTopicActivityName, req, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}
