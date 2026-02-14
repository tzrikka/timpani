package github

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/tzrikka/timpani-api/pkg/github"
	"github.com/tzrikka/timpani/pkg/otel"
)

// UsersGetActivity is based on:
//   - https://docs.github.com/en/rest/users/users?apiVersion=2022-11-28#get-the-authenticated-user
//   - https://docs.github.com/en/rest/users/users?apiVersion=2022-11-28#get-a-user-using-their-id
//   - https://docs.github.com/en/rest/users/users?apiVersion=2022-11-28#get-a-user
func (a *API) UsersGetActivity(ctx context.Context, req github.UsersGetRequest) (map[string]any, error) {
	path := "/user"
	if req.AccountID != "" && req.Username != "" {
		return nil, errors.New("account ID and username are both optional and mutually-exclusive, specify at most one")
	}
	if req.Username != "" {
		path += "s"
	}
	path = fmt.Sprintf("%s/%s%s", path, req.AccountID, req.Username)

	t := time.Now().UTC()
	resp := map[string]any{}
	_, err := a.httpGet(ctx, "", path, nil, &resp)
	otel.IncrementAPICallCounter(t, github.UsersGetActivityName, err)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

// UsersListActivity is based on:
// https://docs.github.com/en/rest/users/users?apiVersion=2022-11-28#list-users
func (a *API) UsersListActivity(ctx context.Context, req github.UsersListRequest) ([]map[string]any, error) {
	query := url.Values{}
	if req.Since != 0 {
		query.Set("since", strconv.Itoa(req.Since))
	}
	if req.PerPage != 0 {
		query.Set("per_page", strconv.Itoa(req.PerPage))
	}

	t := time.Now().UTC()
	resp := []map[string]any{}
	_, err := a.httpGet(ctx, "", "/users/list", query, &resp)
	otel.IncrementAPICallCounter(t, github.UsersListActivityName, err)

	if err != nil {
		return nil, err
	}
	return resp, nil
}
