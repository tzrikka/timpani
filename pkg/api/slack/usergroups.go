package slack

import (
	"context"
	"errors"
	"net/url"

	"github.com/tzrikka/timpani-api/pkg/slack"
)

// UserGroupsListActivity is based on:
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

	resp := new(slack.UserGroupsListResponse)
	if err := a.httpGet(ctx, slack.UserGroupsListActivityName, query, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// UserGroupsUsersListActivity is based on:
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

	resp := new(slack.UserGroupsUsersListResponse)
	if err := a.httpGet(ctx, slack.UserGroupsUsersListActivityName, query, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}
