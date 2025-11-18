package bitbucket

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/tzrikka/timpani-api/pkg/bitbucket"
	"github.com/tzrikka/timpani/pkg/metrics"
)

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-commits/#api-repositories-workspace-repo-slug-diff-spec-get
func (a *API) CommitsDiffActivity(ctx context.Context, req bitbucket.CommitsDiffRequest) (string, error) {
	path := fmt.Sprintf("/repositories/%s/%s/diff/%s", req.Workspace, req.RepoSlug, req.Spec)
	t := time.Now().UTC()

	query := url.Values{}
	if req.Path != "" {
		query.Set("path", req.Path)
	}

	resp, err := a.httpGetText(ctx, req.ThrippyLinkID, path, query)
	if err != nil {
		metrics.IncrementAPICallCounter(t, bitbucket.CommitsDiffActivityName, err)
		return "", err
	}

	metrics.IncrementAPICallCounter(t, bitbucket.CommitsDiffActivityName, nil)
	return resp.String(), nil
}
