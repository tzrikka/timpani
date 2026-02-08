package slack

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	"go.temporal.io/sdk/temporal"

	"github.com/tzrikka/timpani-api/pkg/slack"
)

// UsersConversationsActivity is based on:
// https://docs.slack.dev/reference/methods/users.conversations/
func (a *API) UsersConversationsActivity(ctx context.Context, req slack.UsersConversationsRequest) (*slack.UsersConversationsResponse, error) {
	query := url.Values{}
	if req.Types != "" {
		query.Set("types", req.Types)
	}
	if req.User != "" {
		query.Set("user", req.User)
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

	resp := new(slack.UsersConversationsResponse)
	if err := a.httpGet(ctx, slack.UsersConversationsActivityName, query, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// UsersGetPresenceActivity is based on:
// https://docs.slack.dev/reference/methods/users.getPresence/
func (a *API) UsersGetPresenceActivity(ctx context.Context, req slack.UsersGetPresenceRequest) (*slack.UsersGetPresenceResponse, error) {
	query := url.Values{}
	if req.User != "" {
		query.Set("user", req.User)
	}

	resp := new(slack.UsersGetPresenceResponse)
	if err := a.httpGet(ctx, slack.UsersGetPresenceActivityName, query, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// UsersInfoActivity is based on:
// https://docs.slack.dev/reference/methods/users.info/
func (a *API) UsersInfoActivity(ctx context.Context, req slack.UsersInfoRequest) (*slack.UsersInfoResponse, error) {
	query := url.Values{}
	query.Set("user", req.User)
	if req.IncludeLocale {
		query.Set("include_locale", "true")
	}

	resp := new(slack.UsersInfoResponse)
	if err := a.httpGet(ctx, slack.UsersInfoActivityName, query, resp); err != nil {
		return nil, err
	}

	if resp.Error == "user_not_found" {
		return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, req.User, resp)
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// UsersListActivity is based on:
// https://docs.slack.dev/reference/methods/users.list/
func (a *API) UsersListActivity(ctx context.Context, req slack.UsersListRequest) (*slack.UsersListResponse, error) {
	query := url.Values{}
	if req.IncludeLocale {
		query.Set("include_locale", "true")
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

	resp := new(slack.UsersListResponse)
	if err := a.httpGet(ctx, slack.UsersListActivityName, query, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// UsersLookupByEmailActivity is based on:
// https://docs.slack.dev/reference/methods/users.lookupByEmail/
func (a *API) UsersLookupByEmailActivity(ctx context.Context, req slack.UsersLookupByEmailRequest) (*slack.UsersLookupByEmailResponse, error) {
	query := url.Values{}
	query.Set("email", req.Email)

	resp := new(slack.UsersLookupByEmailResponse)
	if err := a.httpGet(ctx, slack.UsersLookupByEmailActivityName, query, resp); err != nil {
		return nil, err
	}

	switch {
	case resp.Error == "users_not_found":
		return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, req.Email, resp)
	case !resp.OK:
		return nil, errors.New("Slack API error: " + resp.Error)
	default:
		return resp, nil
	}
}

// UsersProfileGetActivity is based on:
// https://docs.slack.dev/reference/methods/users.profile.get/
func (a *API) UsersProfileGetActivity(ctx context.Context, req slack.UsersProfileGetRequest) (*slack.UsersProfileGetResponse, error) {
	query := url.Values{}
	if req.User != "" {
		query.Set("user", req.User)
	}
	if req.IncludeLabels {
		query.Set("include_labels", "true")
	}

	resp := new(slack.UsersProfileGetResponse)
	if err := a.httpGet(ctx, slack.UsersProfileGetActivityName, query, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}
