package bitbucket

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.temporal.io/sdk/temporal"

	"github.com/tzrikka/timpani-api/pkg/bitbucket"
	"github.com/tzrikka/timpani/pkg/metrics"
)

type prCommentBody struct {
	Content prCommentContent       `json:"content"`
	Parent  *prCreateCommentParent `json:"parent,omitempty"`
}

type prCommentContent struct {
	Raw string `json:"raw"`
}

type prCreateCommentParent struct {
	ID int `json:"id"`
}

type prMergeBody struct {
	Type              string `json:"type"`
	Message           string `json:"message,omitempty"`
	MergeStrategy     string `json:"merge_strategy,omitempty"`
	CloseSourceBranch bool   `json:"close_source_branch,omitempty"`
}

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-approve-post
func (a *API) PullRequestsApproveActivity(ctx context.Context, req bitbucket.PullRequestsApproveRequest) error {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/approve", req.Workspace, req.RepoSlug, req.PullRequestID)
	t := time.Now().UTC()

	err := a.httpPost(ctx, req.ThrippyLinkID, path, nil, nil)

	metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsApproveActivityName, err)
	return err
}

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-comments-post
func (a *API) PullRequestsCreateCommentActivity(ctx context.Context, req bitbucket.PullRequestsCreateCommentRequest) (*bitbucket.PullRequestsCreateCommentResponse, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/comments", req.Workspace, req.RepoSlug, req.PullRequestID)
	t := time.Now().UTC()

	body := &prCommentBody{Content: prCommentContent{Raw: req.Markdown}}
	if req.ParentID != "" {
		id, err := strconv.Atoi(req.ParentID)
		if err != nil {
			metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsCreateCommentActivityName, err)
			return nil, temporal.NewNonRetryableApplicationError("invalid parent ID", fmt.Sprintf("%T", err), err, req.ParentID)
		}
		body.Parent = &prCreateCommentParent{ID: id}
	}

	resp := new(bitbucket.PullRequestsCreateCommentResponse)
	if err := a.httpPost(ctx, req.ThrippyLinkID, path, body, resp); err != nil {
		metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsCreateCommentActivityName, err)
		return nil, err
	}

	metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsCreateCommentActivityName, nil)
	return resp, nil
}

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-decline-post
func (a *API) PullRequestsDeclineActivity(ctx context.Context, req bitbucket.PullRequestsDeclineRequest) error {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/decline", req.Workspace, req.RepoSlug, req.PullRequestID)
	t := time.Now().UTC()

	err := a.httpPost(ctx, req.ThrippyLinkID, path, nil, nil)

	metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsDeclineActivityName, err)
	return err
}

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-comments-comment-id-delete
func (a *API) PullRequestsDeleteCommentActivity(ctx context.Context, req bitbucket.PullRequestsDeleteCommentRequest) error {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/comments/%s", req.Workspace, req.RepoSlug, req.PullRequestID, req.CommentID)
	t := time.Now().UTC()

	err := a.httpDelete(ctx, req.ThrippyLinkID, path, url.Values{})

	metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsDeleteCommentActivityName, err)
	return err
}

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-activity-get
func (a *API) PullRequestsListActivityLogActivity(ctx context.Context, req bitbucket.PullRequestsListActivityLogRequest) (*bitbucket.PullRequestsListActivityLogResponse, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/activity", req.Workspace, req.RepoSlug, req.PullRequestID)
	query := paginatedQuery(req.PageLen, req.Page)

	var resp *bitbucket.PullRequestsListActivityLogResponse
	activities := []map[string]any{}
	next := "start"

	for next != "" {
		t := time.Now().UTC()
		resp = new(bitbucket.PullRequestsListActivityLogResponse)
		if err := a.httpGet(ctx, req.ThrippyLinkID, path, query, resp); err != nil {
			metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsListActivityLogActivityName, err)
			return nil, err
		}

		if req.AllPages {
			activities = append(activities, resp.Values...)
			next = resp.Next
			if next != "" {
				u, err := url.Parse(next)
				if err != nil {
					metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsListActivityLogActivityName, err)
					return nil, fmt.Errorf("invalid next page URL %q: %w", next, err)
				}
				path = strings.TrimPrefix(u.Path, "/2.0") // [API.httpRequestPrep] adds this automatically.
				query = u.Query()
			}
		} else {
			next = ""
		}

		metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsListActivityLogActivityName, nil)
	}

	if req.AllPages {
		resp.Values = activities
	}

	return resp, nil
}

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-commits-get
func (a *API) PullRequestsListCommitsActivity(ctx context.Context, req bitbucket.PullRequestsListCommitsRequest) (*bitbucket.PullRequestsListCommitsResponse, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/commits", req.Workspace, req.RepoSlug, req.PullRequestID)
	query := paginatedQuery(req.PageLen, req.Page)

	var resp *bitbucket.PullRequestsListCommitsResponse
	commits := []bitbucket.Commit{}
	next := "start"

	for next != "" {
		t := time.Now().UTC()
		resp = new(bitbucket.PullRequestsListCommitsResponse)
		if err := a.httpGet(ctx, req.ThrippyLinkID, path, query, resp); err != nil {
			metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsListCommitsActivityName, err)
			return nil, err
		}

		if req.AllPages {
			commits = append(commits, resp.Values...)
			next = resp.Next
			if next != "" {
				u, err := url.Parse(next)
				if err != nil {
					metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsListCommitsActivityName, err)
					return nil, fmt.Errorf("invalid next page URL %q: %w", next, err)
				}
				path = strings.TrimPrefix(u.Path, "/2.0") // [API.httpRequestPrep] adds this automatically.
				query = u.Query()
			}
		} else {
			next = ""
		}

		metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsListCommitsActivityName, nil)
	}

	if req.AllPages {
		resp.Values = commits
	}

	return resp, nil
}

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-commit-commit-pullrequests-get
func (a *API) PullRequestsListForCommitActivity(ctx context.Context, req bitbucket.PullRequestsListForCommitRequest) (*bitbucket.PullRequestsListForCommitResponse, error) {
	path := fmt.Sprintf("/repositories/%s/%s/commit/%s/pullrequests", req.Workspace, req.RepoSlug, req.Commit)
	query := paginatedQuery(req.PageLen, req.Page)

	var resp *bitbucket.PullRequestsListForCommitResponse
	prs := []map[string]any{}
	next := "start"

	for next != "" {
		t := time.Now().UTC()
		resp = new(bitbucket.PullRequestsListForCommitResponse)
		if err := a.httpGet(ctx, req.ThrippyLinkID, path, query, resp); err != nil {
			metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsListForCommitActivityName, err)
			return nil, err
		}

		if req.AllPages {
			prs = append(prs, resp.Values...)
			next = resp.Next
			if next != "" {
				u, err := url.Parse(next)
				if err != nil {
					metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsListForCommitActivityName, err)
					return nil, fmt.Errorf("invalid next page URL %q: %w", next, err)
				}
				path = strings.TrimPrefix(u.Path, "/2.0") // [API.httpRequestPrep] adds this automatically.
				query = u.Query()
			}
		} else {
			next = ""
		}

		metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsListForCommitActivityName, nil)
	}

	if req.AllPages {
		resp.Values = prs
	}

	return resp, nil
}

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-merge-post
func (a *API) PullRequestsMergeActivity(ctx context.Context, req bitbucket.PullRequestsMergeRequest) error {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/merge", req.Workspace, req.RepoSlug, req.PullRequestID)
	t := time.Now().UTC()

	body := &prMergeBody{
		Type:              req.Type,
		Message:           req.Message,
		MergeStrategy:     req.MergeStrategy,
		CloseSourceBranch: req.CloseSourceBranch,
	}

	if err := a.httpPost(ctx, req.ThrippyLinkID, path, body, nil); err != nil {
		metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsMergeActivityName, err)
		return err
	}

	metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsMergeActivityName, nil)
	return nil
}

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-approve-delete
func (a *API) PullRequestsUnapproveActivity(ctx context.Context, req bitbucket.PullRequestsUnapproveRequest) error {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/approve", req.Workspace, req.RepoSlug, req.PullRequestID)
	t := time.Now().UTC()

	err := a.httpDelete(ctx, req.ThrippyLinkID, path, url.Values{})

	metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsUnapproveActivityName, err)
	return err
}

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-comments-comment-id-put
func (a *API) PullRequestsUpdateCommentActivity(ctx context.Context, req bitbucket.PullRequestsUpdateCommentRequest) error {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/comments/%s", req.Workspace, req.RepoSlug, req.PullRequestID, req.CommentID)
	t := time.Now().UTC()

	body := &prCommentBody{Content: prCommentContent{Raw: req.Markdown}}
	err := a.httpPut(ctx, req.ThrippyLinkID, path, body, nil)

	metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsUpdateCommentActivityName, err)
	return err
}

func paginatedQuery(pageLen, page string) url.Values {
	query := url.Values{}
	query.Set("pagelen", "100") // Default = 10, but we prefer to minimize number of API calls.

	if pageLen != "" {
		query.Set("pagelen", pageLen)
	}
	if page != "" {
		query.Set("page", page)
	}

	return query
}
