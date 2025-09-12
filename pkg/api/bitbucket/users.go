package bitbucket

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/tzrikka/timpani-api/pkg/bitbucket"
)

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-users/#api-user-get
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-users/#api-users-selected-user-get
func (a *API) UsersGetActivity(ctx context.Context, req bitbucket.UsersGetRequest) (*bitbucket.WorkspacesListMembersResponse, error) {
	if req.AccountID != "" && req.UUID != "" {
		return nil, errors.New("account ID and UUID are both optional and mutually-exclusive, specify at most one")
	}

	path := fmt.Sprintf("/user/%s%s", req.AccountID, req.UUID)
	if req.AccountID != "" || req.UUID != "" {
		path = strings.Replace(path, "user", "users", 1)
	}

	resp := new(bitbucket.WorkspacesListMembersResponse)
	if err := a.httpGet(ctx, "", path, nil, resp); err != nil {
		return nil, err
	}

	return resp, nil
}
