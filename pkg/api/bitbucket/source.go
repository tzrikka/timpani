package bitbucket

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"go.temporal.io/sdk/temporal"

	"github.com/tzrikka/timpani-api/pkg/bitbucket"
	"github.com/tzrikka/timpani/pkg/metrics"
)

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

	t := time.Now().UTC()
	resp, err := a.httpGetText(ctx, req.ThrippyLinkID, path, query)
	metrics.IncrementAPICallCounter(t, bitbucket.SourceGetFileActivityName, err)

	if err != nil {
		if strings.HasPrefix(err.Error(), "404 Not Found") {
			return "", temporal.NewNonRetryableApplicationError(err.Error(), "BitbucketAPIError", err, req)
		}
		return "", err
	}
	return resp.String(), nil
}
