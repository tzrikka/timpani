package slack

import (
	"context"
	"errors"
	"net/url"
	"strconv"
)

const (
	ConversationsArchiveName    = "slack.conversations.archive"
	ConversationsCloseName      = "slack.conversations.close"
	ConversationsCreateName     = "slack.conversations.create"
	ConversationsHistoryName    = "slack.conversations.history"
	ConversationsInfoName       = "slack.conversations.info"
	ConversationsInviteName     = "slack.conversations.invite"
	ConversationsJoinName       = "slack.conversations.join"
	ConversationsKickName       = "slack.conversations.kick"
	ConversationsLeaveName      = "slack.conversations.leave"
	ConversationsListName       = "slack.conversations.list"
	ConversationsMembersName    = "slack.conversations.members"
	ConversationsOpenName       = "slack.conversations.open"
	ConversationsRenameName     = "slack.conversations.rename"
	ConversationsRepliesName    = "slack.conversations.replies"
	ConversationsSetPurposeName = "slack.conversations.setPurpose"
	ConversationsSetTopicName   = "slack.conversations.setTopic"
	ConversationsUnarchiveName  = "slack.conversations.unarchive"
)

// https://docs.slack.dev/reference/methods/conversations.archive
type ConversationsArchiveRequest struct {
	Channel string `json:"channel"`
}

// https://docs.slack.dev/reference/methods/conversations.archive
type ConversationsArchiveResponse struct {
	slackResponse
}

// https://docs.slack.dev/reference/methods/conversations.archive
func (a *API) ConversationsArchiveActivity(ctx context.Context, req *ConversationsArchiveRequest) (*ConversationsArchiveResponse, error) {
	resp := new(ConversationsArchiveResponse)
	if err := a.httpPost(ctx, ConversationsArchiveName, req, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.close
type ConversationsCloseRequest struct {
	Channel string `json:"channel"`
}

// https://docs.slack.dev/reference/methods/conversations.close
type ConversationsCloseResponse struct {
	slackResponse

	AlreadyClosed bool `json:"already_closed,omitempty"`
}

// https://docs.slack.dev/reference/methods/conversations.close
func (a *API) ConversationsCloseActivity(ctx context.Context, req *ConversationsCloseRequest) (*ConversationsCloseResponse, error) {
	resp := new(ConversationsCloseResponse)
	if err := a.httpPost(ctx, ConversationsCloseName, req, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.create
type ConversationsCreateRequest struct {
	Name string `json:"name"`

	IsPrivate bool   `json:"is_private,omitempty"`
	TeamID    string `json:"team_id,omitempty"`
}

// https://docs.slack.dev/reference/methods/conversations.create
type ConversationsCreateResponse struct {
	slackResponse

	Channel map[string]any `json:"channel,omitempty"`
}

// https://docs.slack.dev/reference/methods/conversations.create
func (a *API) ConversationsCreateActivity(ctx context.Context, req *ConversationsCreateRequest) (*ConversationsCreateResponse, error) {
	resp := new(ConversationsCreateResponse)
	if err := a.httpPost(ctx, ConversationsCreateName, req, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.history
type ConversationsHistoryRequest struct {
	Channel string `json:"channel"`

	Cursor             string `json:"cursor,omitempty"`
	IncludeAllMetadata bool   `json:"include_all_metadata,omitempty"`
	Inclusive          bool   `json:"inclusive,omitempty"`
	Latest             string `json:"latest,omitempty"`
	Limit              int    `json:"limit,omitempty"`
	Oldest             string `json:"oldest,omitempty"`
}

// https://docs.slack.dev/reference/methods/conversations.history
type ConversationsHistoryResponse struct {
	slackResponse

	Latest    string           `json:"latest,omitempty"`
	Messages  []map[string]any `json:"messages,omitempty"`
	HasMore   bool             `json:"has_more,omitempty"`
	IsLimited bool             `json:"is_limited,omitempty"` // Undocumented.
	PinCount  int              `json:"pin_count,omitempty"`
	// Undocumented: "channel_actions_ts" and "channel_actions_count".
}

// https://docs.slack.dev/reference/methods/conversations.history
func (a *API) ConversationsHistoryActivity(ctx context.Context, req *ConversationsHistoryRequest) (*ConversationsHistoryResponse, error) {
	query := url.Values{}
	query.Set("channel", req.Channel)
	if req.Cursor != "" {
		query.Set("cursor", req.Cursor)
	}
	if req.IncludeAllMetadata {
		query.Set("include_all_metadata", "true")
	}
	if req.Inclusive {
		query.Set("inclusive", "true")
	}
	if req.Latest != "" {
		query.Set("latest", req.Latest)
	}
	if req.Limit != 0 {
		query.Set("limit", strconv.Itoa(req.Limit))
	}
	if req.Oldest != "" {
		query.Set("oldest", req.Oldest)
	}

	resp := new(ConversationsHistoryResponse)
	if err := a.httpGet(ctx, ConversationsHistoryName, query, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.info
type ConversationsInfoRequest struct {
	Channel string `json:"channel"`

	IncludeLocale     bool `json:"include_locale,omitempty"`
	IncludeNumMembers bool `json:"include_num_members,omitempty"`
}

// https://docs.slack.dev/reference/methods/conversations.info
type ConversationsInfoResponse struct {
	slackResponse

	Channel map[string]any `json:"channel,omitempty"`
}

// https://docs.slack.dev/reference/methods/conversations.info
func (a *API) ConversationsInfoActivity(ctx context.Context, req *ConversationsInfoRequest) (*ConversationsInfoResponse, error) {
	query := url.Values{}
	query.Set("channel", req.Channel)
	if req.IncludeLocale {
		query.Set("include_locale", "true")
	}
	if req.IncludeNumMembers {
		query.Set("include_num_members", "true")
	}

	resp := new(ConversationsInfoResponse)
	if err := a.httpGet(ctx, ConversationsInfoName, query, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.invite
type ConversationsInviteRequest struct {
	Channel string `json:"channel"`
	Users   string `json:"users"`

	Force bool `json:"force,omitempty"`
}

// https://docs.slack.dev/reference/methods/conversations.invite
type ConversationsInviteResponse struct {
	slackResponse

	Channel map[string]any   `json:"channel,omitempty"`
	Errors  []map[string]any `json:"errors,omitempty"`
}

// https://docs.slack.dev/reference/methods/conversations.invite
func (a *API) ConversationsInviteActivity(ctx context.Context, req *ConversationsInviteRequest) (*ConversationsInviteResponse, error) {
	resp := new(ConversationsInviteResponse)
	if err := a.httpPost(ctx, ConversationsInviteName, req, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.join
type ConversationsJoinRequest struct {
	Channel string `json:"channel"`
}

// https://docs.slack.dev/reference/methods/conversations.join
type ConversationsJoinResponse struct {
	slackResponse

	Channel map[string]any `json:"channel,omitempty"`
}

// https://docs.slack.dev/reference/methods/conversations.join
func (a *API) ConversationsJoinActivity(ctx context.Context, req *ConversationsJoinRequest) (*ConversationsJoinResponse, error) {
	resp := new(ConversationsJoinResponse)
	if err := a.httpPost(ctx, ConversationsJoinName, req, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.kick
type ConversationsKickRequest struct {
	Channel string `json:"channel"`

	User string `json:"user,omitempty"`
}

// https://docs.slack.dev/reference/methods/conversations.kick
type ConversationsKickResponse struct {
	slackResponse
}

// https://docs.slack.dev/reference/methods/conversations.kick
func (a *API) ConversationsKickActivity(ctx context.Context, req *ConversationsKickRequest) (*ConversationsKickResponse, error) {
	resp := new(ConversationsKickResponse)
	if err := a.httpPost(ctx, ConversationsKickName, req, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.leave
type ConversationsLeaveRequest struct {
	Channel string `json:"channel"`
}

// https://docs.slack.dev/reference/methods/conversations.leave
type ConversationsLeaveResponse struct {
	slackResponse

	NotInChannel bool `json:"not_in_channel,omitempty"`
}

// https://docs.slack.dev/reference/methods/conversations.leave
func (a *API) ConversationsLeaveActivity(ctx context.Context, req *ConversationsLeaveRequest) (*ConversationsLeaveResponse, error) {
	resp := new(ConversationsLeaveResponse)
	if err := a.httpPost(ctx, ConversationsLeaveName, req, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.list
type ConversationsListRequest struct {
	Cursor          string `json:"cursor,omitempty"`
	ExcludeArchived bool   `json:"exclude_archived,omitempty"`
	Limit           int    `json:"limit,omitempty"`
	TeamID          string `json:"team_id,omitempty"`
	Types           string `json:"types,omitempty"`
}

// https://docs.slack.dev/reference/methods/conversations.list
type ConversationsListResponse struct {
	slackResponse

	Channels []map[string]any `json:"channels,omitempty"`
}

// https://docs.slack.dev/reference/methods/conversations.list
func (a *API) ConversationsListActivity(ctx context.Context, req *ConversationsListRequest) (*ConversationsListResponse, error) {
	query := url.Values{}
	if req.Cursor != "" {
		query.Set("cursor", req.Cursor)
	}
	if req.ExcludeArchived {
		query.Set("exclude_archived", "true")
	}
	if req.Limit != 0 {
		query.Set("limit", strconv.Itoa(req.Limit))
	}
	if req.TeamID != "" {
		query.Set("team_id", req.TeamID)
	}
	if req.Types != "" {
		query.Set("types", req.Types)
	}

	resp := new(ConversationsListResponse)
	if err := a.httpGet(ctx, ConversationsListName, query, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.members
type ConversationsMembersRequest struct {
	Channel string `json:"channel"`

	Cursor string `json:"cursor,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

// https://docs.slack.dev/reference/methods/conversations.members
type ConversationsMembersResponse struct {
	slackResponse

	Members []string `json:"members,omitempty"`
}

// https://docs.slack.dev/reference/methods/conversations.members
func (a *API) ConversationsMembersActivity(ctx context.Context, req *ConversationsMembersRequest) (*ConversationsMembersResponse, error) {
	query := url.Values{}
	query.Set("channel", req.Channel)
	if req.Cursor != "" {
		query.Set("cursor", req.Cursor)
	}
	if req.Limit != 0 {
		query.Set("limit", strconv.Itoa(req.Limit))
	}

	resp := new(ConversationsMembersResponse)
	if err := a.httpGet(ctx, ConversationsMembersName, query, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.open
type ConversationsOpenRequest struct {
	Channel         string `json:"channel,omitempty"`
	ReturnIM        bool   `json:"return_im,omitempty"`
	Users           string `json:"users,omitempty"`
	PreventCreation bool   `json:"prevent_creation,omitempty"`
}

// https://docs.slack.dev/reference/methods/conversations.open
type ConversationsOpenResponse struct {
	slackResponse

	NoOp        bool           `json:"no_op,omitempty"`
	AlreadyOpen bool           `json:"already_open,omitempty"`
	Channel     map[string]any `json:"channel,omitempty"`
}

// https://docs.slack.dev/reference/methods/conversations.open
func (a *API) ConversationsOpenActivity(ctx context.Context, req *ConversationsOpenRequest) (*ConversationsOpenResponse, error) {
	resp := new(ConversationsOpenResponse)
	if err := a.httpPost(ctx, ConversationsOpenName, req, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.rename
type ConversationsRenameRequest struct {
	Channel string `json:"channel"`
	Name    string `json:"name"`
}

// https://docs.slack.dev/reference/methods/conversations.rename
type ConversationsRenameResponse struct {
	slackResponse

	Channel map[string]any `json:"channel,omitempty"`
}

// https://docs.slack.dev/reference/methods/conversations.rename
func (a *API) ConversationsRenameActivity(ctx context.Context, req *ConversationsRenameRequest) (*ConversationsRenameResponse, error) {
	resp := new(ConversationsRenameResponse)
	if err := a.httpPost(ctx, ConversationsRenameName, req, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.replies
type ConversationsRepliesRequest struct {
	Channel string `json:"channel"`
	TS      string `json:"ts"`

	Cursor             string `json:"cursor,omitempty"`
	IncludeAllMetadata bool   `json:"include_all_metadata,omitempty"`
	Inclusive          bool   `json:"inclusive,omitempty"`
	Latest             string `json:"latest,omitempty"`
	Limit              int    `json:"limit,omitempty"`
	Oldest             string `json:"oldest,omitempty"`
}

// https://docs.slack.dev/reference/methods/conversations.replies
type ConversationsRepliesResponse struct {
	slackResponse

	Messages []map[string]any `json:"messages,omitempty"`
	HasMore  bool             `json:"has_more,omitempty"`
}

// https://docs.slack.dev/reference/methods/conversations.replies
func (a *API) ConversationsRepliesActivity(ctx context.Context, req *ConversationsRepliesRequest) (*ConversationsRepliesResponse, error) {
	query := url.Values{}
	query.Set("channel", req.Channel)
	query.Set("ts", req.TS)
	if req.Cursor != "" {
		query.Set("cursor", req.Cursor)
	}
	if req.IncludeAllMetadata {
		query.Set("include_all_metadata", "true")
	}
	if req.Inclusive {
		query.Set("inclusive", "true")
	}
	if req.Latest != "" {
		query.Set("latest", req.Latest)
	}
	if req.Limit != 0 {
		query.Set("limit", strconv.Itoa(req.Limit))
	}
	if req.Oldest != "" {
		query.Set("oldest", req.Oldest)
	}

	resp := new(ConversationsRepliesResponse)
	if err := a.httpGet(ctx, ConversationsRepliesName, query, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.setPurpose
type ConversationsSetPurposeRequest struct {
	Channel string `json:"channel"`
	Purpose string `json:"purpose"`
}

// https://docs.slack.dev/reference/methods/conversations.setPurpose
type ConversationsSetPurposeResponse struct {
	slackResponse

	Channel map[string]any `json:"channel,omitempty"` // Empirically different from the documentation.
}

// https://docs.slack.dev/reference/methods/conversations.setPurpose
func (a *API) ConversationsSetPurposeActivity(ctx context.Context, req *ConversationsSetPurposeRequest) (*ConversationsSetPurposeResponse, error) {
	resp := new(ConversationsSetPurposeResponse)
	if err := a.httpPost(ctx, ConversationsSetPurposeName, req, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.setTopic
type ConversationsSetTopicRequest struct {
	Channel string `json:"channel"`
	Topic   string `json:"topic"`
}

// https://docs.slack.dev/reference/methods/conversations.setTopic
type ConversationsSetTopicResponse struct {
	slackResponse

	Channel map[string]any `json:"channel,omitempty"`
}

// https://docs.slack.dev/reference/methods/conversations.setTopic
func (a *API) ConversationsSetTopicActivity(ctx context.Context, req *ConversationsSetTopicRequest) (*ConversationsSetTopicResponse, error) {
	resp := new(ConversationsSetTopicResponse)
	if err := a.httpPost(ctx, ConversationsSetTopicName, req, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/conversations.unarchive
type ConversationsUnarchiveRequest struct {
	Channel string `json:"channel"`
}

// https://docs.slack.dev/reference/methods/conversations.unarchive
type ConversationsUnarchiveResponse struct {
	slackResponse
}

// https://docs.slack.dev/reference/methods/conversations.unarchive
func (a *API) ConversationsUnarchiveActivity(ctx context.Context, req *ConversationsUnarchiveRequest) (*ConversationsUnarchiveResponse, error) {
	resp := new(ConversationsUnarchiveResponse)
	if err := a.httpPost(ctx, ConversationsUnarchiveName, req, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}
