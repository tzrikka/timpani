package slack

import (
	"context"
	"errors"
	"net/url"
	"strconv"
)

const (
	UsersConversationsName = "slack.users.conversations"
	UsersGetPresenceName   = "slack.users.getPresence"
	UsersIdentityName      = "slack.users.identity"
	UsersInfoName          = "slack.users.info"
	UsersListName          = "slack.users.list"
	UsersLookupByEmailName = "slack.users.lookupByEmail"
	UsersProfileGetName    = "slack.users.profile.get"
)

// https://docs.slack.dev/reference/methods/users.conversations
type UsersConversationsRequest struct {
	Cursor          string `json:"cursor,omitempty"`
	ExcludeArchived bool   `json:"exclude_archived,omitempty"`
	Limit           int    `json:"limit,omitempty"`
	TeamID          string `json:"team_id,omitempty"`
	Types           string `json:"types,omitempty"`
	User            string `json:"user,omitempty"`
}

// https://docs.slack.dev/reference/methods/users.conversations
type UsersConversationsResponse struct {
	slackResponse

	Channels []map[string]any `json:"channels,omitempty"`
}

// https://docs.slack.dev/reference/methods/users.conversations
func (a *API) UsersConversationsActivity(ctx context.Context, req UsersConversationsRequest) (*UsersConversationsResponse, error) {
	query := url.Values{}
	if req.Cursor != "" {
		query.Set("cursor", req.Cursor)
	}
	if req.ExcludeArchived {
		query.Set("exclude_archived", "true")
	}
	if req.Limit != 0 {
		query.Set("limit", strconv.Itoa(req.Limit))
	}
	if req.TeamID != "" {
		query.Set("team_id", req.TeamID)
	}
	if req.Types != "" {
		query.Set("types", req.Types)
	}
	if req.User != "" {
		query.Set("user", req.User)
	}

	resp := new(UsersConversationsResponse)
	if err := a.httpGet(ctx, UsersConversationsName, query, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/users.getPresence
type UsersGetPresenceRequest struct {
	User string `json:"user,omitempty"`
}

// https://docs.slack.dev/reference/methods/users.getPresence
type UsersGetPresenceResponse struct {
	slackResponse

	Presence        string `json:"presence,omitempty"`
	Online          bool   `json:"online,omitempty"`
	AutoAway        bool   `json:"auto_away,omitempty"`
	ManualAway      bool   `json:"manual_away,omitempty"`
	ConnectionCount int    `json:"connection_count,omitempty"`
	LastActivity    int64  `json:"last_activity,omitempty"`
}

// https://docs.slack.dev/reference/methods/users.getPresence
func (a *API) UsersGetPresenceActivity(ctx context.Context, req UsersGetPresenceRequest) (*UsersGetPresenceResponse, error) {
	query := url.Values{}
	if req.User != "" {
		query.Set("user", req.User)
	}

	resp := new(UsersGetPresenceResponse)
	if err := a.httpGet(ctx, UsersGetPresenceName, query, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/users.identity
type UsersIdentityRequest struct{}

// https://docs.slack.dev/reference/methods/users.identity
type UsersIdentityResponse struct {
	slackResponse

	User map[string]any `json:"user,omitempty"`
	Team map[string]any `json:"team,omitempty"`
}

// https://docs.slack.dev/reference/methods/users.identity
func (a *API) UsersIdentityActivity(ctx context.Context, _ UsersIdentityRequest) (*UsersIdentityResponse, error) {
	resp := new(UsersIdentityResponse)
	if err := a.httpGet(ctx, UsersIdentityName, url.Values{}, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/users.info
type UsersInfoRequest struct {
	User string `json:"user"`

	IncludeLocale bool `json:"include_locale,omitempty"`
}

// https://docs.slack.dev/reference/methods/users.info
type UsersInfoResponse struct {
	slackResponse

	User map[string]any `json:"user,omitempty"`
}

// https://docs.slack.dev/reference/methods/users.info
func (a *API) UsersInfoActivity(ctx context.Context, req UsersInfoRequest) (*UsersInfoResponse, error) {
	query := url.Values{}
	query.Set("user", req.User)
	if req.IncludeLocale {
		query.Set("include_locale", "true")
	}

	resp := new(UsersInfoResponse)
	if err := a.httpGet(ctx, UsersInfoName, query, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/users.list
type UsersListRequest struct {
	Cursor        string `json:"cursor,omitempty"`
	IncludeLocale bool   `json:"include_locale,omitempty"`
	Limit         int    `json:"limit,omitempty"`
	TeamID        string `json:"team_id,omitempty"`
}

// https://docs.slack.dev/reference/methods/users.list
type UsersListResponse struct {
	slackResponse

	Members []map[string]any `json:"members,omitempty"`
	CacheTS int64            `json:"cache_ts,omitempty"`
}

// https://docs.slack.dev/reference/methods/users.list
func (a *API) UsersListActivity(ctx context.Context, req UsersListRequest) (*UsersListResponse, error) {
	query := url.Values{}
	if req.Cursor != "" {
		query.Set("cursor", req.Cursor)
	}
	if req.IncludeLocale {
		query.Set("include_locale", "true")
	}
	if req.Limit != 0 {
		query.Set("limit", strconv.Itoa(req.Limit))
	}
	if req.TeamID != "" {
		query.Set("team_id", req.TeamID)
	}

	resp := new(UsersListResponse)
	if err := a.httpGet(ctx, UsersListName, query, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/users.lookupByEmail
type UsersLookupByEmailRequest struct {
	Email string `json:"email"`
}

// https://docs.slack.dev/reference/methods/users.lookupByEmail
type UsersLookupByEmailResponse struct {
	slackResponse

	User map[string]any `json:"user,omitempty"`
}

// https://docs.slack.dev/reference/methods/users.lookupByEmail
func (a *API) UsersLookupByEmailActivity(ctx context.Context, req UsersLookupByEmailRequest) (*UsersLookupByEmailResponse, error) {
	query := url.Values{}
	query.Set("email", req.Email)

	resp := new(UsersLookupByEmailResponse)
	if err := a.httpGet(ctx, UsersLookupByEmailName, query, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/users.profile.get
type UsersProfileGetRequest struct {
	IncludeLabels bool   `json:"include_labels,omitempty"`
	User          string `json:"user,omitempty"`
}

// https://docs.slack.dev/reference/methods/users.profile.get
type UsersProfileGetResponse struct {
	slackResponse

	Profile map[string]any `json:"profile,omitempty"`
}

// https://docs.slack.dev/reference/methods/users.profile.get
func (a *API) UsersProfileGetActivity(ctx context.Context, req UsersProfileGetRequest) (*UsersProfileGetResponse, error) {
	query := url.Values{}
	if req.IncludeLabels {
		query.Set("include_labels", "true")
	}
	if req.User != "" {
		query.Set("user", req.User)
	}

	resp := new(UsersProfileGetResponse)
	if err := a.httpGet(ctx, UsersProfileGetName, query, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}
