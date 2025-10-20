package bitbucket

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"go.temporal.io/sdk/temporal"

	"github.com/tzrikka/timpani-api/pkg/bitbucket"
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

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-comments-post
func (a *API) PullRequestsCreateCommentActivity(ctx context.Context, req bitbucket.PullRequestsCreateCommentRequest) (*bitbucket.PullRequestsCreateCommentResponse, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/comments", req.Workspace, req.RepoSlug, req.PullRequestID)

	body := &prCommentBody{Content: prCommentContent{Raw: req.Markdown}}
	if req.ParentID != "" {
		id, err := strconv.Atoi(req.ParentID)
		if err != nil {
			return nil, temporal.NewNonRetryableApplicationError("invalid parent ID", fmt.Sprintf("%T", err), err, req.ParentID)
		}
		body.Parent = &prCreateCommentParent{ID: id}
	}

	resp := new(bitbucket.PullRequestsCreateCommentResponse)
	if err := a.httpPost(ctx, req.ThrippyLinkID, path, body, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-comments-comment-id-delete
func (a *API) PullRequestsDeleteCommentActivity(ctx context.Context, req bitbucket.PullRequestsDeleteCommentRequest) error {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/comments/%s", req.Workspace, req.RepoSlug, req.PullRequestID, req.CommentID)
	return a.httpDelete(ctx, req.ThrippyLinkID, path, url.Values{})
}

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-commits-get
func (a *API) PullRequestsListCommitsActivity(ctx context.Context, req bitbucket.PullRequestsListCommitsRequest) (*bitbucket.PullRequestsListCommitsResponse, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/commits", req.Workspace, req.RepoSlug, req.PullRequestID)

	query := url.Values{}
	query.Set("pagelen", "100") // Default = 10, but we prefer to minimize number of API calls.
	if req.PageLen != "" {
		query.Set("pagelen", req.PageLen)
	}
	if req.Page != "" {
		query.Set("page", req.Page)
	}

	var resp *bitbucket.PullRequestsListCommitsResponse
	commits := []bitbucket.Commit{}
	next := "start"

	for next != "" {
		resp = new(bitbucket.PullRequestsListCommitsResponse)
		if err := a.httpGet(ctx, req.ThrippyLinkID, path, query, resp); err != nil {
			return nil, err
		}

		if req.AllPages {
			commits = append(commits, resp.Values...)
			next = resp.Next
			if next != "" {
				u, err := url.Parse(next)
				if err != nil {
					return nil, fmt.Errorf("invalid next page URL %q: %w", next, err)
				}
				path = u.Path
				query = u.Query()
			}
		}
	}

	if req.AllPages {
		resp.Values = commits
	}

	return resp, nil
}

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-comments-comment-id-put
func (a *API) PullRequestsUpdateCommentActivity(ctx context.Context, req bitbucket.PullRequestsUpdateCommentRequest) error {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/comments/%s", req.Workspace, req.RepoSlug, req.PullRequestID, req.CommentID)
	body := &prCommentBody{Content: prCommentContent{Raw: req.Markdown}}
	return a.httpPut(ctx, req.ThrippyLinkID, path, body, nil)
}
