package bitbucket

import (
	"context"
	"fmt"

	"github.com/tzrikka/timpani-api/pkg/bitbucket"
)

type prCreateCommentBody struct {
	Content prCreateCommentContent `json:"content"`
}

type prCreateCommentContent struct {
	Raw string `json:"raw"`
}

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-comments-post
func (a *API) PullRequestsCreateCommentActivity(ctx context.Context, req bitbucket.PullRequestsCreateCommentRequest) (*bitbucket.PullRequestsCreateCommentResponse, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/comments", req.Workspace, req.RepoSlug, req.PullRequestID)

	body := &prCreateCommentBody{
		Content: prCreateCommentContent{
			Raw: req.Markdown,
		},
	}

	resp := new(bitbucket.PullRequestsCreateCommentResponse)
	if err := a.httpPost(ctx, req.ThrippyLinkID, path, body, resp); err != nil {
		return nil, err
	}

	return resp, nil
}
