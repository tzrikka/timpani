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

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-comments-comment-id-put
func (a *API) PullRequestsUpdateCommentActivity(ctx context.Context, req bitbucket.PullRequestsUpdateCommentRequest) error {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/comments/%s", req.Workspace, req.RepoSlug, req.PullRequestID, req.CommentID)
	body := &prCommentBody{Content: prCommentContent{Raw: req.Markdown}}
	return a.httpPut(ctx, req.ThrippyLinkID, path, body, nil)
}
