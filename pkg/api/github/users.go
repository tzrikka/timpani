package github

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

const (
	UsersGetName  = "github.users.get"
	UsersListName = "github.users.list"
)

// https://docs.github.com/en/rest/users/users?apiVersion=2022-11-28#get-the-authenticated-user
// https://docs.github.com/en/rest/users/users?apiVersion=2022-11-28#get-a-user-using-their-id
// https://docs.github.com/en/rest/users/users?apiVersion=2022-11-28#get-a-user
type UsersGetRequest struct {
	AccountID string `json:"account_id,omitempty"`
	Username  string `json:"username,omitempty"`
}

// https://docs.github.com/en/rest/users/users?apiVersion=2022-11-28#get-the-authenticated-user
// https://docs.github.com/en/rest/users/users?apiVersion=2022-11-28#get-a-user-using-their-id
// https://docs.github.com/en/rest/users/users?apiVersion=2022-11-28#get-a-user
func (a *API) UsersGetActivity(ctx context.Context, req UsersGetRequest) (map[string]any, error) {
	path := "/user"
	if req.AccountID != "" && req.Username != "" {
		return nil, errors.New("account ID and username are both optional and mutually-exclusive, specify at most one")
	}
	if req.Username != "" {
		path += "s"
	}
	path = fmt.Sprintf("%s/%s%s", path, req.AccountID, req.Username)

	resp := map[string]any{}
	err := a.httpGet(ctx, path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// https://docs.github.com/en/rest/users/users?apiVersion=2022-11-28#list-users
type UsersListRequest struct {
	Since   int `json:"since,omitempty"`
	PerPage int `json:"per_page,omitempty"`
}

// https://docs.github.com/en/rest/users/users?apiVersion=2022-11-28#list-users
func (a *API) UsersListActivity(ctx context.Context, req UsersListRequest) ([]map[string]any, error) {
	query := url.Values{}
	if req.Since != 0 {
		query.Set("since", strconv.Itoa(req.Since))
	}
	if req.PerPage != 0 {
		query.Set("per_page", strconv.Itoa(req.PerPage))
	}

	resp := []map[string]any{}
	err := a.httpGet(ctx, "/users/list", query, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
