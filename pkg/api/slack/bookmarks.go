package slack

import (
	"context"
	"errors"

	"github.com/tzrikka/timpani-api/pkg/slack"
)

// https://docs.slack.dev/reference/methods/bookmarks.add/
func (a *API) BookmarksAddActivity(ctx context.Context, req slack.BookmarksAddRequest) (*slack.BookmarksAddResponse, error) {
	resp := new(slack.BookmarksAddResponse)
	if err := a.httpPost(ctx, slack.BookmarksAddActivityName, req, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	return resp, nil
}

// https://docs.slack.dev/reference/methods/bookmarks.edit/
func (a *API) BookmarksEditActivity(ctx context.Context, req slack.BookmarksEditRequest) (*slack.BookmarksEditResponse, error) {
	resp := new(slack.BookmarksEditResponse)
	if err := a.httpPost(ctx, slack.BookmarksEditActivityName, req, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	return resp, nil
}

// https://docs.slack.dev/reference/methods/bookmarks.list/
func (a *API) BookmarksListActivity(ctx context.Context, req slack.BookmarksListRequest) (*slack.BookmarksListResponse, error) {
	resp := new(slack.BookmarksListResponse)
	if err := a.httpPost(ctx, slack.BookmarksListActivityName, req, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	return resp, nil
}

// https://docs.slack.dev/reference/methods/bookmarks.remove/
func (a *API) BookmarksRemoveActivity(ctx context.Context, req slack.BookmarksRemoveRequest) (*slack.BookmarksRemoveResponse, error) {
	resp := new(slack.BookmarksRemoveResponse)
	if err := a.httpPost(ctx, slack.BookmarksRemoveActivityName, req, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	return resp, nil
}
