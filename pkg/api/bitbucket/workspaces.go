package bitbucket

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/tzrikka/timpani-api/pkg/bitbucket"
)

// WorkspacesListMembersActivity is based on:
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-workspaces/#api-workspaces-workspace-members-get
func (a *API) WorkspacesListMembersActivity(
	ctx context.Context,
	req bitbucket.WorkspacesListMembersRequest,
) (*bitbucket.WorkspacesListMembersResponse, error) {
	path := fmt.Sprintf("/workspaces/%s/members", req.Workspace)
	query := url.Values{}
	if len(req.EmailsFilter) > 0 {
		query.Set("q", fmt.Sprintf(`user.email IN ("%s")`, strings.Join(req.EmailsFilter, `","`)))
	}

	resp := new(bitbucket.WorkspacesListMembersResponse)
	err := a.httpGet(ctx, bitbucket.WorkspacesListMembersActivityName, "", path, query, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
