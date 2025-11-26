package bitbucket

import (
	"context"
	"fmt"
	"net/url"
	"strings"
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

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-commits/#api-repositories-workspace-repo-slug-diffstat-spec-get
func (a *API) CommitsDiffStatActivity(ctx context.Context, req bitbucket.CommitsDiffStatRequest) (*bitbucket.CommitsDiffStatResponse, error) {
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

	var resp *bitbucket.CommitsDiffStatResponse
	diffstats := []bitbucket.DiffStat{}
	next := "start"

	for next != "" {
		t := time.Now().UTC()
		resp = new(bitbucket.CommitsDiffStatResponse)
		if err := a.httpGet(ctx, req.ThrippyLinkID, path, query, resp); err != nil {
			metrics.IncrementAPICallCounter(t, bitbucket.CommitsDiffStatActivityName, err)
			return nil, err
		}

		if req.AllPages {
			diffstats = append(diffstats, resp.Values...)
			next = resp.Next
			if next != "" {
				u, err := url.Parse(next)
				if err != nil {
					metrics.IncrementAPICallCounter(t, bitbucket.CommitsDiffStatActivityName, err)
					return nil, fmt.Errorf("invalid next page URL %q: %w", next, err)
				}
				path = strings.TrimPrefix(u.Path, "/2.0") // [API.httpRequestPrep] adds this automatically.
				query = u.Query()
			}
		} else {
			next = ""
		}

		metrics.IncrementAPICallCounter(t, bitbucket.CommitsDiffStatActivityName, nil)
	}

	if req.AllPages {
		resp.Values = diffstats
	}

	return resp, nil
}
