package slack

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	"github.com/tzrikka/timpani-api/pkg/slack"
)

// https://docs.slack.dev/reference/methods/reactions.add/
func (a *API) ReactionsAddActivity(ctx context.Context, req slack.ReactionsAddRequest) (*slack.ReactionsAddResponse, error) {
	resp := new(slack.ReactionsAddResponse)
	if err := a.httpPost(ctx, slack.ReactionsAddActivityName, req, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	return resp, nil
}

// https://docs.slack.dev/reference/methods/reactions.get/
func (a *API) ReactionsGetActivity(ctx context.Context, req slack.ReactionsGetRequest) (*slack.ReactionsGetResponse, error) {
	query := url.Values{}
	if req.Channel != "" {
		query.Set("channel", req.Channel)
	}
	if req.Timestamp != "" {
		query.Set("timestamp", req.Timestamp)
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

	resp := new(slack.ReactionsGetResponse)
	if err := a.httpGet(ctx, slack.ReactionsGetActivityName, query, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	return resp, nil
}

// https://docs.slack.dev/reference/methods/reactions.list/
func (a *API) ReactionsListActivity(ctx context.Context, req slack.ReactionsListRequest) (*slack.ReactionsListResponse, error) {
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
	if req.Limit != 0 {
		query.Set("limit", strconv.Itoa(req.Limit))
	}
	if req.Cursor != "" {
		query.Set("cursor", req.Cursor)
	}
	if req.TeamID != "" {
		query.Set("team_id", req.TeamID)
	}

	resp := new(slack.ReactionsListResponse)
	if err := a.httpGet(ctx, slack.ReactionsListActivityName, query, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	return resp, nil
}

// https://docs.slack.dev/reference/methods/reactions.remove/
func (a *API) ReactionsRemoveActivity(ctx context.Context, req slack.ReactionsRemoveRequest) (*slack.ReactionsRemoveResponse, error) {
	resp := new(slack.ReactionsRemoveResponse)
	if err := a.httpPost(ctx, slack.ReactionsRemoveActivityName, req, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	return resp, nil
}
