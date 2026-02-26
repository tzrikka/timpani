package bitbucket

import (
	"context"
	"fmt"
	"net/url"

	"github.com/tzrikka/timpani-api/pkg/bitbucket"
)

// CommitsDiffActivity is based on:
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-commits/#api-repositories-workspace-repo-slug-diff-spec-get
func (a *API) CommitsDiffActivity(ctx context.Context, req bitbucket.CommitsDiffRequest) (string, error) {
	path := fmt.Sprintf("/repositories/%s/%s/diff/%s", req.Workspace, req.RepoSlug, req.Spec)
	query := url.Values{}
	if req.Path != "" {
		query.Set("path", req.Path)
	}

	return a.httpGetText(ctx, bitbucket.CommitsDiffActivityName, req.ThrippyLinkID, path, query)
}

// CommitsDiffstatActivity is based on:
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-commits/#api-repositories-workspace-repo-slug-diffstat-spec-get
func (a *API) CommitsDiffstatActivity(ctx context.Context, req bitbucket.CommitsDiffstatRequest) (*bitbucket.CommitsDiffstatResponse, error) {
	path := fmt.Sprintf("/repositories/%s/%s/diffstat/%s", req.Workspace, req.RepoSlug, req.Spec)
	path, query, err := paginatedQuery(bitbucket.CommitsDiffstatActivityName, path, req.PageLen, req.Page, req.Next)
	if err != nil {
		return nil, err
	}

	if req.IgnoreWhitespace {
		query.Set("ignore_whitespace", "true")
	}
	if req.Merge {
		query.Set("merge", "true")
	}
	if req.Renames {
		query.Set("renames", "true")
	}
	if req.Topic {
		query.Set("topic", "true")
	}
	if req.Path != "" {
		query.Set("path", req.Path)
	}

	resp := new(bitbucket.CommitsDiffstatResponse)
	err = a.httpGet(ctx, bitbucket.CommitsDiffstatActivityName, req.ThrippyLinkID, path, query, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
