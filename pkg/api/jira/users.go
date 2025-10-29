package jira

import (
	"context"
	"net/url"
	"time"

	"github.com/tzrikka/timpani-api/pkg/jira"
	"github.com/tzrikka/timpani/pkg/metrics"
)

// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-users/#api-rest-api-3-user-get
func (a *API) UsersGetActivity(ctx context.Context, req jira.UsersGetRequest) (*jira.UsersGetResponse, error) {
	t := time.Now().UTC()
	query := url.Values{}
	query.Set("accountId", req.AccountID)

	resp := new(jira.UsersGetResponse)
	if err := a.httpGet(ctx, "user", query, resp); err != nil {
		metrics.CountAPICall(t, jira.UsersGetActivityName, err)
		return nil, err
	}

	metrics.CountAPICall(t, jira.UsersGetActivityName, nil)
	return resp, nil
}

// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-user-search/#api-rest-api-3-user-search-get
func (a *API) UsersSearchActivity(ctx context.Context, req jira.UsersSearchRequest) ([]jira.User, error) {
	t := time.Now().UTC()
	query := url.Values{}
	query.Set("query", req.Query)

	var resp []jira.User
	if err := a.httpGet(ctx, "user/search", query, &resp); err != nil {
		metrics.CountAPICall(t, jira.UsersSearchActivityName, err)
		return nil, err
	}

	metrics.CountAPICall(t, jira.UsersSearchActivityName, nil)
	return resp, nil
}
