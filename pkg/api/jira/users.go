package jira

import (
	"context"
	"net/url"
	"time"

	"github.com/tzrikka/timpani-api/pkg/jira"
	"github.com/tzrikka/timpani/pkg/otel"
)

// UsersGetActivity is based on:
// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-users/#api-rest-api-3-user-get
func (a *API) UsersGetActivity(ctx context.Context, req jira.UsersGetRequest) (*jira.UsersGetResponse, error) {
	query := url.Values{}
	query.Set("accountId", req.AccountID)

	t := time.Now().UTC()
	resp := new(jira.UsersGetResponse)
	err := a.httpGet(ctx, "user", query, resp)
	otel.IncrementAPICallCounter(t, jira.UsersGetActivityName, err)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

// UsersSearchActivity is based on:
// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-user-search/#api-rest-api-3-user-search-get
func (a *API) UsersSearchActivity(ctx context.Context, req jira.UsersSearchRequest) ([]jira.User, error) {
	query := url.Values{}
	query.Set("query", req.Query)

	t := time.Now().UTC()
	var resp []jira.User
	err := a.httpGet(ctx, "user/search", query, &resp)
	otel.IncrementAPICallCounter(t, jira.UsersSearchActivityName, err)

	if err != nil {
		return nil, err
	}
	return resp, nil
}
