package slack

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/tzrikka/timpani-api/pkg/slack"
	"github.com/tzrikka/timpani/pkg/metrics"
)

// ReactionsAddActivity is based on:
// https://docs.slack.dev/reference/methods/reactions.add/
func (a *API) ReactionsAddActivity(ctx context.Context, req slack.ReactionsAddRequest) (*slack.ReactionsAddResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.ReactionsAddResponse)
	if err := a.httpPost(ctx, slack.ReactionsAddActivityName, req, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.ReactionsAddActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.ReactionsAddActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.ReactionsAddActivityName, nil)
	return resp, nil
}

// ReactionsGetActivity is based on:
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

	t := time.Now().UTC()
	resp := new(slack.ReactionsGetResponse)
	if err := a.httpGet(ctx, slack.ReactionsGetActivityName, query, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.ReactionsGetActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.ReactionsGetActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.ReactionsGetActivityName, nil)
	return resp, nil
}

// ReactionsListActivity is based on:
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

	t := time.Now().UTC()
	resp := new(slack.ReactionsListResponse)
	if err := a.httpGet(ctx, slack.ReactionsListActivityName, query, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.ReactionsListActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.ReactionsListActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.ReactionsListActivityName, nil)
	return resp, nil
}

// ReactionsRemoveActivity is based on:
// https://docs.slack.dev/reference/methods/reactions.remove/
func (a *API) ReactionsRemoveActivity(ctx context.Context, req slack.ReactionsRemoveRequest) (*slack.ReactionsRemoveResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.ReactionsRemoveResponse)
	if err := a.httpPost(ctx, slack.ReactionsRemoveActivityName, req, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.ReactionsRemoveActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.ReactionsRemoveActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.ReactionsRemoveActivityName, nil)
	return resp, nil
}
