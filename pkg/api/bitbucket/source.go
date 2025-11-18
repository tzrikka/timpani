package bitbucket

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/tzrikka/timpani-api/pkg/bitbucket"
	"github.com/tzrikka/timpani/pkg/metrics"
)

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-source/#api-repositories-workspace-repo-slug-src-commit-path-get
func (a *API) SourceGetFileActivity(ctx context.Context, req bitbucket.SourceGetRequest) (string, error) {
	path := fmt.Sprintf("/repositories/%s/%s/src/%s/%s", req.Workspace, req.RepoSlug, req.Commit, req.Path)
	t := time.Now().UTC()

	query := url.Values{}
	if req.Filter != "" {
		query.Set("q", req.Filter)
	}
	if req.Sort != "" {
		query.Set("sort", req.Sort)
	}

	resp, err := a.httpGetText(ctx, req.ThrippyLinkID, path, query)
	if err != nil {
		metrics.IncrementAPICallCounter(t, bitbucket.SourceGetFileActivityName, err)
		return "", err
	}

	metrics.IncrementAPICallCounter(t, bitbucket.SourceGetFileActivityName, nil)
	return resp.String(), nil
}
