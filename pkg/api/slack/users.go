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

	t := time.Now().UTC()
	resp := new(slack.UsersConversationsResponse)
	if err := a.httpGet(ctx, slack.UsersConversationsActivityName, query, resp); err != nil {
		metrics.CountAPICall(t, slack.UsersConversationsActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.CountAPICall(t, slack.UsersConversationsActivityName, errors.New(resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.CountAPICall(t, slack.UsersConversationsActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/users.getPresence/
func (a *API) UsersGetPresenceActivity(ctx context.Context, req slack.UsersGetPresenceRequest) (*slack.UsersGetPresenceResponse, error) {
	query := url.Values{}
	if req.User != "" {
		query.Set("user", req.User)
	}

	t := time.Now().UTC()
	resp := new(slack.UsersGetPresenceResponse)
	if err := a.httpGet(ctx, slack.UsersGetPresenceActivityName, query, resp); err != nil {
		metrics.CountAPICall(t, slack.UsersGetPresenceActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.CountAPICall(t, slack.UsersGetPresenceActivityName, errors.New(resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.CountAPICall(t, slack.UsersGetPresenceActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/users.info/
func (a *API) UsersInfoActivity(ctx context.Context, req slack.UsersInfoRequest) (*slack.UsersInfoResponse, error) {
	query := url.Values{}
	query.Set("user", req.User)
	if req.IncludeLocale {
		query.Set("include_locale", "true")
	}

	t := time.Now().UTC()
	resp := new(slack.UsersInfoResponse)
	if err := a.httpGet(ctx, slack.UsersInfoActivityName, query, resp); err != nil {
		metrics.CountAPICall(t, slack.UsersInfoActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.CountAPICall(t, slack.UsersInfoActivityName, errors.New(resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.CountAPICall(t, slack.UsersInfoActivityName, nil)
	return resp, nil
}

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

	t := time.Now().UTC()
	resp := new(slack.UsersListResponse)
	if err := a.httpGet(ctx, slack.UsersListActivityName, query, resp); err != nil {
		metrics.CountAPICall(t, slack.UsersListActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.CountAPICall(t, slack.UsersListActivityName, errors.New(resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.CountAPICall(t, slack.UsersListActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/users.lookupByEmail/
func (a *API) UsersLookupByEmailActivity(ctx context.Context, req slack.UsersLookupByEmailRequest) (*slack.UsersLookupByEmailResponse, error) {
	query := url.Values{}
	query.Set("email", req.Email)

	t := time.Now().UTC()
	resp := new(slack.UsersLookupByEmailResponse)
	if err := a.httpGet(ctx, slack.UsersLookupByEmailActivityName, query, resp); err != nil {
		metrics.CountAPICall(t, slack.UsersLookupByEmailActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.CountAPICall(t, slack.UsersLookupByEmailActivityName, errors.New(resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.CountAPICall(t, slack.UsersLookupByEmailActivityName, nil)
	return resp, nil
}

// https://docs.slack.dev/reference/methods/users.profile.get/
func (a *API) UsersProfileGetActivity(ctx context.Context, req slack.UsersProfileGetRequest) (*slack.UsersProfileGetResponse, error) {
	query := url.Values{}
	if req.User != "" {
		query.Set("user", req.User)
	}
	if req.IncludeLabels {
		query.Set("include_labels", "true")
	}

	t := time.Now().UTC()
	resp := new(slack.UsersProfileGetResponse)
	if err := a.httpGet(ctx, slack.UsersProfileGetActivityName, query, resp); err != nil {
		metrics.CountAPICall(t, slack.UsersProfileGetActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.CountAPICall(t, slack.UsersProfileGetActivityName, errors.New(resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.CountAPICall(t, slack.UsersProfileGetActivityName, nil)
	return resp, nil
}
