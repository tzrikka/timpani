package github

import (
	"context"
	"fmt"
	"time"

	"github.com/tzrikka/timpani-api/pkg/github"
	"github.com/tzrikka/timpani/pkg/otel"
)

type issueCommentMarkdown struct {
	Body string `json:"body"`
}

// IssuesCommentsCreateActivity is based on:
// https://docs.github.com/en/rest/issues/comments?apiVersion=2022-11-28#create-an-issue-comment
func (a *API) IssuesCommentsCreateActivity(ctx context.Context, req github.IssuesCommentsCreateRequest) (*github.IssueComment, error) {
	path := fmt.Sprintf("/repos/%s/%s/issues/%d/comments", req.Owner, req.Repo, req.IssueNumber)

	t := time.Now().UTC()
	resp := new(github.IssueComment)
	err := a.httpPost(ctx, req.ThrippyLinkID, path, "application/vnd.github.raw+json", issueCommentMarkdown{Body: req.Body}, resp)
	otel.IncrementAPICallCounter(t, github.IssuesCommentsCreateActivityName, err)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

// IssuesCommentsDeleteActivity is based on:
// https://docs.github.com/en/rest/issues/comments?apiVersion=2022-11-28#delete-an-issue-comment
func (a *API) IssuesCommentsDeleteActivity(ctx context.Context, req github.IssuesCommentsDeleteRequest) error {
	path := fmt.Sprintf("/repos/%s/%s/issues/comments/%d", req.Owner, req.Repo, req.CommentID)

	t := time.Now().UTC()
	err := a.httpDelete(ctx, req.ThrippyLinkID, path, nil)
	otel.IncrementAPICallCounter(t, github.IssuesCommentsDeleteActivityName, err)

	if err != nil {
		return err
	}
	return nil
}

// IssuesCommentsUpdateActivity is based on:
// https://docs.github.com/en/rest/issues/comments?apiVersion=2022-11-28#update-an-issue-comment
func (a *API) IssuesCommentsUpdateActivity(ctx context.Context, req github.IssuesCommentsUpdateRequest) (*github.IssueComment, error) {
	path := fmt.Sprintf("/repos/%s/%s/issues/comments/%d", req.Owner, req.Repo, req.CommentID)

	t := time.Now().UTC()
	resp := new(github.IssueComment)
	err := a.httpPatch(ctx, req.ThrippyLinkID, path, "application/vnd.github.raw+json", issueCommentMarkdown{Body: req.Body}, resp)
	otel.IncrementAPICallCounter(t, github.IssuesCommentsUpdateActivityName, err)

	if err != nil {
		return nil, err
	}
	return resp, nil
}
