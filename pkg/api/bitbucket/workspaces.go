package bitbucket

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/tzrikka/timpani-api/pkg/bitbucket"
	"github.com/tzrikka/timpani/pkg/otel"
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

	t := time.Now().UTC()
	resp := new(bitbucket.WorkspacesListMembersResponse)
	err := a.httpGet(ctx, "", path, query, resp)
	otel.IncrementAPICallCounter(t, bitbucket.WorkspacesListMembersActivityName, err)

	if err != nil {
		return nil, err
	}
	return resp, nil
}
