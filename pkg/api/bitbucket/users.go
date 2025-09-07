package bitbucket

import (
	"context"
	"errors"
	"fmt"

	"github.com/tzrikka/timpani-api/pkg/bitbucket"
)

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-users/#api-user-get
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-users/#api-users-selected-user-get
func (a *API) UsersGetActivity(ctx context.Context, req bitbucket.UsersGetRequest) (*bitbucket.WorkspacesListMembersResponse, error) {
	path := "/user"
	if req.AccountID != "" && req.UUID != "" {
		return nil, errors.New("account ID and UUID are both optional and mutually-exclusive, specify at most one")
	}
	if req.AccountID != "" || req.UUID != "" {
		path += "s"
	}
	path = fmt.Sprintf("%s/%s%s", path, req.AccountID, req.UUID)

	resp := new(bitbucket.WorkspacesListMembersResponse)
	if err := a.httpGet(ctx, path, nil, resp); err != nil {
		return nil, err
	}

	return resp, nil
}
