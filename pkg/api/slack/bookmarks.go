package slack

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.temporal.io/sdk/temporal"

	"github.com/tzrikka/timpani-api/pkg/slack"
	"github.com/tzrikka/timpani/pkg/metrics"
)

// BookmarksAddActivity is based on:
// https://docs.slack.dev/reference/methods/bookmarks.add/
func (a *API) BookmarksAddActivity(ctx context.Context, req slack.BookmarksAddRequest) (*slack.BookmarksAddResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.BookmarksAddResponse)
	if err := a.httpPost(ctx, slack.BookmarksAddActivityName, req, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.BookmarksAddActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.BookmarksAddActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.BookmarksAddActivityName, nil)
	return resp, nil
}

// BookmarksEditActivity is based on:
// https://docs.slack.dev/reference/methods/bookmarks.edit/
func (a *API) BookmarksEditActivity(ctx context.Context, req slack.BookmarksEditRequest) (*slack.BookmarksEditResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.BookmarksEditResponse)
	if err := a.httpPost(ctx, slack.BookmarksEditActivityName, req, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.BookmarksEditActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.BookmarksEditActivityName, slackAPIError(resp, resp.Error))

		if resp.Error == "permission_denied" {
			return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, req.ChannelID, resp)
		}
		if strings.Contains(resp.Error, "invalid") {
			return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, resp)
		}
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.BookmarksEditActivityName, nil)
	return resp, nil
}

// BookmarksListActivity is based on:
// https://docs.slack.dev/reference/methods/bookmarks.list/
func (a *API) BookmarksListActivity(ctx context.Context, req slack.BookmarksListRequest) (*slack.BookmarksListResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.BookmarksListResponse)
	if err := a.httpPost(ctx, slack.BookmarksListActivityName, req, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.BookmarksListActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.BookmarksListActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.BookmarksListActivityName, nil)
	return resp, nil
}

// BookmarksRemoveActivity is based on:
// https://docs.slack.dev/reference/methods/bookmarks.remove/
func (a *API) BookmarksRemoveActivity(ctx context.Context, req slack.BookmarksRemoveRequest) (*slack.BookmarksRemoveResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.BookmarksRemoveResponse)
	if err := a.httpPost(ctx, slack.BookmarksRemoveActivityName, req, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.BookmarksRemoveActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.BookmarksRemoveActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.BookmarksRemoveActivityName, nil)
	return resp, nil
}
