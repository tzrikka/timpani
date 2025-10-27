package slack

import (
	"context"
	"errors"
	"time"

	"github.com/tzrikka/timpani-api/pkg/slack"
	"github.com/tzrikka/timpani/pkg/metrics"
)

// https://docs.slack.dev/reference/methods/bookmarks.add/
func (a *API) BookmarksAddActivity(ctx context.Context, req slack.BookmarksAddRequest) (*slack.BookmarksAddResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.BookmarksAddResponse)
	if err := a.httpPost(ctx, slack.BookmarksAddActivityName, req, resp); err != nil {
		metrics.CountAPICall(t, slack.BookmarksAddActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.CountAPICall(t, slack.BookmarksAddActivityName, errors.New(resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.CountAPICall(t, slack.BookmarksAddActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/bookmarks.edit/
func (a *API) BookmarksEditActivity(ctx context.Context, req slack.BookmarksEditRequest) (*slack.BookmarksEditResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.BookmarksEditResponse)
	if err := a.httpPost(ctx, slack.BookmarksEditActivityName, req, resp); err != nil {
		metrics.CountAPICall(t, slack.BookmarksEditActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.CountAPICall(t, slack.BookmarksEditActivityName, errors.New(resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.CountAPICall(t, slack.BookmarksEditActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/bookmarks.list/
func (a *API) BookmarksListActivity(ctx context.Context, req slack.BookmarksListRequest) (*slack.BookmarksListResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.BookmarksListResponse)
	if err := a.httpPost(ctx, slack.BookmarksListActivityName, req, resp); err != nil {
		metrics.CountAPICall(t, slack.BookmarksListActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.CountAPICall(t, slack.BookmarksListActivityName, errors.New(resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.CountAPICall(t, slack.BookmarksListActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/bookmarks.remove/
func (a *API) BookmarksRemoveActivity(ctx context.Context, req slack.BookmarksRemoveRequest) (*slack.BookmarksRemoveResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.BookmarksRemoveResponse)
	if err := a.httpPost(ctx, slack.BookmarksRemoveActivityName, req, resp); err != nil {
		metrics.CountAPICall(t, slack.BookmarksRemoveActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.CountAPICall(t, slack.BookmarksRemoveActivityName, errors.New(resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.CountAPICall(t, slack.BookmarksRemoveActivityName, nil)
	return resp, nil
}
