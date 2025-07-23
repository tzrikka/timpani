package slack

import (
	"context"
	"errors"
	"net/url"
	"strconv"
)

const (
	ReactionsAddName    = "slack.reactions.add"
	ReactionsGetName    = "slack.reactions.get"
	ReactionsListName   = "slack.reactions.list"
	ReactionsRemoveName = "slack.reactions.remove"
)

// https://docs.slack.dev/reference/methods/reactions.add
type ReactionsAddRequest struct {
	Channel   string `json:"channel"`
	Name      string `json:"name"`
	Timestamp string `json:"timestamp"`
}

// https://docs.slack.dev/reference/methods/reactions.add
type ReactionsAddResponse struct {
	slackResponse
}

// https://docs.slack.dev/reference/methods/reactions.add
func (a *API) ReactionsAddActivity(ctx context.Context, req *ReactionsAddRequest) (*ReactionsAddResponse, error) {
	resp := new(ReactionsAddResponse)
	if err := a.httpPost(ctx, ReactionsAddName, req, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/reactions.get
type ReactionsGetRequest struct {
	Channel     string `json:"channel,omitempty"`
	File        string `json:"file,omitempty"`
	FileComment string `json:"file_comment,omitempty"`
	Full        bool   `json:"full,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
}

// https://docs.slack.dev/reference/methods/reactions.get
type ReactionsGetResponse struct {
	slackResponse

	Type    string         `json:"type,omitempty"`
	Message map[string]any `json:"message,omitempty"`
	Channel string         `json:"channel,omitempty"`
}

// https://docs.slack.dev/reference/methods/reactions.get
func (a *API) ReactionsGetActivity(ctx context.Context, req *ReactionsGetRequest) (*ReactionsGetResponse, error) {
	query := url.Values{}
	if req.Channel != "" {
		query.Set("channel", req.Channel)
	}
	if req.File != "" {
		query.Set("file", req.File)
	}
	if req.FileComment != "" {
		query.Set("file_comment", req.FileComment)
	}
	if req.Full {
		query.Set("full", "true")
	}
	if req.Timestamp != "" {
		query.Set("timestamp", req.Timestamp)
	}

	resp := new(ReactionsGetResponse)
	if err := a.httpGet(ctx, ReactionsGetName, query, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/reactions.list
type ReactionsListRequest struct {
	User   string `json:"user,omitempty"`
	Full   bool   `json:"full,omitempty"`
	Count  int    `json:"count,omitempty"`
	Page   int    `json:"page,omitempty"`
	Cursor string `json:"cursor,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	TeamID string `json:"team_id,omitempty"`
}

// https://docs.slack.dev/reference/methods/reactions.list
type ReactionsListResponse struct {
	slackResponse

	Items []map[string]any `json:"items,omitempty"`
}

// https://docs.slack.dev/reference/methods/reactions.list
func (a *API) ReactionsListActivity(ctx context.Context, req *ReactionsListRequest) (*ReactionsListResponse, error) {
	query := url.Values{}
	if req.User != "" {
		query.Set("user", req.User)
	}
	if req.Full {
		query.Set("full", "true")
	}
	if req.Count != 0 {
		query.Set("count", strconv.Itoa(req.Count))
	}
	if req.Page != 0 {
		query.Set("page", strconv.Itoa(req.Page))
	}
	if req.Cursor != "" {
		query.Set("cursor", req.Cursor)
	}
	if req.Limit != 0 {
		query.Set("limit", strconv.Itoa(req.Limit))
	}
	if req.TeamID != "" {
		query.Set("team_id", req.TeamID)
	}

	resp := new(ReactionsListResponse)
	if err := a.httpGet(ctx, ReactionsListName, query, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/reactions.remove
type ReactionsRemoveRequest struct {
	Name string `json:"name"`

	Channel     string `json:"channel,omitempty"`
	File        string `json:"file,omitempty"`
	FileComment string `json:"file_comment,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
}

// https://docs.slack.dev/reference/methods/reactions.remove
type ReactionsRemoveResponse struct {
	slackResponse
}

// https://docs.slack.dev/reference/methods/reactions.remove
func (a *API) ReactionsRemoveActivity(ctx context.Context, req *ReactionsRemoveRequest) (*ReactionsRemoveResponse, error) {
	resp := new(ReactionsRemoveResponse)
	if err := a.httpPost(ctx, ReactionsRemoveName, req, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}
