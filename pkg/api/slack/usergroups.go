package slack

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/tzrikka/timpani-api/pkg/slack"
	"github.com/tzrikka/timpani/pkg/metrics"
)

// https://docs.slack.dev/reference/methods/usergroups.list/
func (a *API) UserGroupsListActivity(ctx context.Context, req slack.UserGroupsListRequest) (*slack.UserGroupsListResponse, error) {
	query := url.Values{}
	if req.IncludeCount {
		query.Set("include_count", "true")
	}
	if req.IncludeDisabled {
		query.Set("include_disabled", "true")
	}
	if req.IncludeUsers {
		query.Set("include_users", "true")
	}
	if req.TeamID != "" {
		query.Set("team_id", req.TeamID)
	}

	t := time.Now().UTC()
	resp := new(slack.UserGroupsListResponse)
	if err := a.httpGet(ctx, slack.UserGroupsListActivityName, query, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.UserGroupsListActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.UserGroupsListActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.UserGroupsListActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/usergroups.users.list
func (a *API) UserGroupsUsersListActivity(ctx context.Context, req slack.UserGroupsUsersListRequest) (*slack.UserGroupsUsersListResponse, error) {
	query := url.Values{}
	query.Set("usergroup", req.Usergroup)
	if req.IncludeDisabled {
		query.Set("include_disabled", "true")
	}
	if req.TeamID != "" {
		query.Set("team_id", req.TeamID)
	}

	t := time.Now().UTC()
	resp := new(slack.UserGroupsUsersListResponse)
	if err := a.httpGet(ctx, slack.UserGroupsUsersListActivityName, query, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.UserGroupsUsersListActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.UserGroupsUsersListActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.UserGroupsUsersListActivityName, nil)
	return resp, nil
}
