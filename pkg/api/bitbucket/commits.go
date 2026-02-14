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

// CommitsDiffActivity is based on:
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-commits/#api-repositories-workspace-repo-slug-diff-spec-get
func (a *API) CommitsDiffActivity(ctx context.Context, req bitbucket.CommitsDiffRequest) (string, error) {
	path := fmt.Sprintf("/repositories/%s/%s/diff/%s", req.Workspace, req.RepoSlug, req.Spec)

	query := url.Values{}
	if req.Path != "" {
		query.Set("path", req.Path)
	}

	t := time.Now().UTC()
	resp, err := a.httpGetText(ctx, req.ThrippyLinkID, path, query)
	otel.IncrementAPICallCounter(t, bitbucket.CommitsDiffActivityName, err)

	if err != nil {
		return "", err
	}
	return resp.String(), nil
}

// CommitsDiffstatActivity is based on:
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-commits/#api-repositories-workspace-repo-slug-diffstat-spec-get
func (a *API) CommitsDiffstatActivity(ctx context.Context, req bitbucket.CommitsDiffstatRequest) (*bitbucket.CommitsDiffstatResponse, error) {
	path := fmt.Sprintf("/repositories/%s/%s/diffstat/%s", req.Workspace, req.RepoSlug, req.Spec)

	query := paginatedQuery(req.PageLen, req.Page)
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

	t := time.Now().UTC()
	if req.Next != "" {
		overrideURL, err := url.Parse(req.Next)
		if err != nil {
			err = fmt.Errorf("invalid next page URL %q: %w", req.Next, err)
			otel.IncrementAPICallCounter(t, bitbucket.CommitsDiffstatActivityName, err)
			return nil, err
		}

		path = strings.TrimPrefix(overrideURL.Path, "/2.0") // [API.httpRequestPrep] adds this automatically.
		query = overrideURL.Query()
	}

	resp := new(bitbucket.CommitsDiffstatResponse)
	err := a.httpGet(ctx, req.ThrippyLinkID, path, query, resp)
	otel.IncrementAPICallCounter(t, bitbucket.CommitsDiffstatActivityName, err)

	if err != nil {
		return nil, err
	}
	return resp, nil
}
