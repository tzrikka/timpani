package bitbucket

import (
	"context"
	"fmt"
	"net/url"

	"github.com/tzrikka/timpani-api/pkg/bitbucket"
)

// SourceGetFileActivity is based on:
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-source/#api-repositories-workspace-repo-slug-src-commit-path-get
func (a *API) SourceGetFileActivity(ctx context.Context, req bitbucket.SourceGetRequest) (string, error) {
	path := fmt.Sprintf("/repositories/%s/%s/src/%s/%s", req.Workspace, req.RepoSlug, req.Commit, req.Path)
	query := url.Values{}
	if req.Filter != "" {
		query.Set("q", req.Filter)
	}
	if req.Sort != "" {
		query.Set("sort", req.Sort)
	}

	return a.httpGetText(ctx, bitbucket.SourceGetFileActivityName, req.ThrippyLinkID, path, query)
}
