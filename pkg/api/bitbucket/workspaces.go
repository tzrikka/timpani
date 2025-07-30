package bitbucket

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

const (
	WorkspacesListMembersName = "bitbucket.workspaces.listMembers"
)

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-workspaces/#api-workspaces-workspace-members-get
type WorkspacesListMembersRequest struct {
	EmailsFilter []string `json:"emails_filter"`
}

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-workspaces/#api-workspaces-workspace-members-get
func (a *API) WorkspacesListMembersActivity(ctx context.Context, req WorkspacesListMembersRequest) (map[string]any, error) {
	path := "/workspaces/TODO/members"

	query := url.Values{}
	if len(req.EmailsFilter) > 0 {
		query.Set("q", fmt.Sprintf(`user.email IN ("%s")`, strings.Join(req.EmailsFilter, `","`)))
	}

	resp := map[string]any{}
	if err := a.httpGet(ctx, path, query, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
