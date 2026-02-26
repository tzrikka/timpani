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
	"github.com/tzrikka/timpani/pkg/otel"
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

// PullRequestsApproveActivity is based on:
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-approve-post
func (a *API) PullRequestsApproveActivity(ctx context.Context, req bitbucket.PullRequestsApproveRequest) error {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/approve", req.Workspace, req.RepoSlug, req.PullRequestID)

	t := time.Now().UTC()
	err := a.httpPost(ctx, req.ThrippyLinkID, path, nil, nil)
	otel.IncrementAPICallCounter(t, bitbucket.PullRequestsApproveActivityName, err)

	return err
}

// PullRequestsCreateCommentActivity is based on:
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-comments-post
func (a *API) PullRequestsCreateCommentActivity(
	ctx context.Context,
	req bitbucket.PullRequestsCreateCommentRequest,
) (*bitbucket.PullRequestsCreateCommentResponse, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/comments", req.Workspace, req.RepoSlug, req.PullRequestID)

	t := time.Now().UTC()
	body := &prCommentBody{Content: prCommentContent{Raw: req.Markdown}}
	if req.ParentID != "" {
		id, err := strconv.Atoi(req.ParentID)
		if err != nil {
			otel.IncrementAPICallCounter(t, bitbucket.PullRequestsCreateCommentActivityName, err)
			return nil, temporal.NewNonRetryableApplicationError("invalid parent ID", fmt.Sprintf("%T", err), err, req.ParentID)
		}
		body.Parent = &prCreateCommentParent{ID: id}
	}

	resp := new(bitbucket.PullRequestsCreateCommentResponse)
	err := a.httpPost(ctx, req.ThrippyLinkID, path, body, resp)
	otel.IncrementAPICallCounter(t, bitbucket.PullRequestsCreateCommentActivityName, err)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

// PullRequestsDeclineActivity is based on:
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-decline-post
func (a *API) PullRequestsDeclineActivity(ctx context.Context, req bitbucket.PullRequestsDeclineRequest) error {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/decline", req.Workspace, req.RepoSlug, req.PullRequestID)

	t := time.Now().UTC()
	err := a.httpPost(ctx, req.ThrippyLinkID, path, nil, nil)
	otel.IncrementAPICallCounter(t, bitbucket.PullRequestsDeclineActivityName, err)

	return err
}

// PullRequestsDeleteCommentActivity is based on:
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-comments-comment-id-delete
func (a *API) PullRequestsDeleteCommentActivity(ctx context.Context, req bitbucket.PullRequestsDeleteCommentRequest) error {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/comments/%s", req.Workspace, req.RepoSlug, req.PullRequestID, req.CommentID)

	t := time.Now().UTC()
	err := a.httpDelete(ctx, req.ThrippyLinkID, path, url.Values{})
	otel.IncrementAPICallCounter(t, bitbucket.PullRequestsDeleteCommentActivityName, err)

	return err
}

// PullRequestsDiffstatActivity is based on:
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-diffstat-get
func (a *API) PullRequestsDiffstatActivity(
	ctx context.Context,
	req bitbucket.PullRequestsDiffstatRequest,
) (*bitbucket.PullRequestsDiffstatResponse, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/diffstat", req.Workspace, req.RepoSlug, req.PullRequestID)
	path, query, err := paginatedQuery(bitbucket.PullRequestsDiffstatActivityName, path, req.PageLen, req.Page, req.Next)
	if err != nil {
		return nil, err
	}

	resp := new(bitbucket.PullRequestsDiffstatResponse)
	err = a.httpGet(ctx, bitbucket.PullRequestsDiffstatActivityName, req.ThrippyLinkID, path, query, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// PullRequestsGetActivity is based on:
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-get
func (a *API) PullRequestsGetActivity(ctx context.Context, req bitbucket.PullRequestsGetRequest) (map[string]any, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s", req.Workspace, req.RepoSlug, req.PullRequestID)

	resp := map[string]any{}
	err := a.httpGet(ctx, bitbucket.PullRequestsGetActivityName, req.ThrippyLinkID, path, url.Values{}, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// PullRequestsGetCommentActivity is based on:
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-comments-comment-id-get
func (a *API) PullRequestsGetCommentActivity(
	ctx context.Context,
	req bitbucket.PullRequestsGetCommentRequest,
) (*bitbucket.PullRequestsGetCommentResponse, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/comments/%s", req.Workspace, req.RepoSlug, req.PullRequestID, req.CommentID)

	resp := new(bitbucket.PullRequestsGetCommentResponse)
	err := a.httpGet(ctx, bitbucket.PullRequestsGetCommentActivityName, req.ThrippyLinkID, path, url.Values{}, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// PullRequestsListActivityLogActivity is based on:
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-activity-get
func (a *API) PullRequestsListActivityLogActivity(
	ctx context.Context,
	req bitbucket.PullRequestsListActivityLogRequest,
) (*bitbucket.PullRequestsListActivityLogResponse, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/activity", req.Workspace, req.RepoSlug, req.PullRequestID)
	path, query, err := paginatedQuery(bitbucket.PullRequestsListActivityLogActivityName, path, req.PageLen, req.Page, req.Next)
	if err != nil {
		return nil, err
	}

	resp := new(bitbucket.PullRequestsListActivityLogResponse)
	err = a.httpGet(ctx, bitbucket.PullRequestsListActivityLogActivityName, req.ThrippyLinkID, path, query, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// PullRequestsListCommitsActivity is based on:
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-commits-get
func (a *API) PullRequestsListCommitsActivity(
	ctx context.Context,
	req bitbucket.PullRequestsListCommitsRequest,
) (*bitbucket.PullRequestsListCommitsResponse, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/commits", req.Workspace, req.RepoSlug, req.PullRequestID)
	path, query, err := paginatedQuery(bitbucket.PullRequestsListCommitsActivityName, path, req.PageLen, req.Page, req.Next)
	if err != nil {
		return nil, err
	}

	resp := new(bitbucket.PullRequestsListCommitsResponse)
	err = a.httpGet(ctx, bitbucket.PullRequestsListCommitsActivityName, req.ThrippyLinkID, path, query, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// PullRequestsListForCommitActivity is based on:
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-commit-commit-pullrequests-get
func (a *API) PullRequestsListForCommitActivity(
	ctx context.Context,
	req bitbucket.PullRequestsListForCommitRequest,
) (*bitbucket.PullRequestsListForCommitResponse, error) {
	path := fmt.Sprintf("/repositories/%s/%s/commit/%s/pullrequests", req.Workspace, req.RepoSlug, req.Commit)
	path, query, err := paginatedQuery(bitbucket.PullRequestsListForCommitActivityName, path, req.PageLen, req.Page, req.Next)
	if err != nil {
		return nil, err
	}

	resp := new(bitbucket.PullRequestsListForCommitResponse)
	err = a.httpGet(ctx, bitbucket.PullRequestsListForCommitActivityName, req.ThrippyLinkID, path, query, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// PullRequestsListTasksActivity is based on:
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-tasks-get
func (a *API) PullRequestsListTasksActivity(
	ctx context.Context,
	req bitbucket.PullRequestsListTasksRequest,
) (*bitbucket.PullRequestsListTasksResponse, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/tasks", req.Workspace, req.RepoSlug, req.PullRequestID)
	path, query, err := paginatedQuery(bitbucket.PullRequestsListTasksActivityName, path, req.PageLen, req.Page, req.Next)
	if err != nil {
		return nil, err
	}

	resp := new(bitbucket.PullRequestsListTasksResponse)
	err = a.httpGet(ctx, bitbucket.PullRequestsListTasksActivityName, req.ThrippyLinkID, path, query, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// PullRequestsMergeActivity is based on:
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-merge-post
func (a *API) PullRequestsMergeActivity(ctx context.Context, req bitbucket.PullRequestsMergeRequest) error {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/merge", req.Workspace, req.RepoSlug, req.PullRequestID)
	body := &prMergeBody{
		Type:              req.Type,
		Message:           req.Message,
		MergeStrategy:     req.MergeStrategy,
		CloseSourceBranch: req.CloseSourceBranch,
	}

	t := time.Now().UTC()
	err := a.httpPost(ctx, req.ThrippyLinkID, path, body, nil)
	otel.IncrementAPICallCounter(t, bitbucket.PullRequestsMergeActivityName, err)

	return err
}

// PullRequestsUnapproveActivity is based on:
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-approve-delete
func (a *API) PullRequestsUnapproveActivity(ctx context.Context, req bitbucket.PullRequestsUnapproveRequest) error {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/approve", req.Workspace, req.RepoSlug, req.PullRequestID)

	t := time.Now().UTC()
	err := a.httpDelete(ctx, req.ThrippyLinkID, path, url.Values{})
	otel.IncrementAPICallCounter(t, bitbucket.PullRequestsUnapproveActivityName, err)

	return err
}

// PullRequestsUpdateActivity is based on:
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-put
func (a *API) PullRequestsUpdateActivity(ctx context.Context, req bitbucket.PullRequestsUpdateRequest) (map[string]any, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s", req.Workspace, req.RepoSlug, req.PullRequestID)

	t := time.Now().UTC()
	resp := map[string]any{}
	err := a.httpPut(ctx, req.ThrippyLinkID, path, req.PullRequest, &resp)
	otel.IncrementAPICallCounter(t, bitbucket.PullRequestsUpdateActivityName, err)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

// PullRequestsUpdateCommentActivity is based on:
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-comments-comment-id-put
func (a *API) PullRequestsUpdateCommentActivity(ctx context.Context, req bitbucket.PullRequestsUpdateCommentRequest) error {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/comments/%s", req.Workspace, req.RepoSlug, req.PullRequestID, req.CommentID)
	body := &prCommentBody{Content: prCommentContent{Raw: req.Markdown}}

	t := time.Now().UTC()
	err := a.httpPut(ctx, req.ThrippyLinkID, path, body, nil)
	otel.IncrementAPICallCounter(t, bitbucket.PullRequestsUpdateCommentActivityName, err)

	return err
}

func paginatedQuery(name, path, pageLen, page, next string) (string, url.Values, error) {
	query := url.Values{}
	query.Set("pagelen", "100") // Default = 10, but we prefer to minimize the number of API calls.

	if pageLen != "" {
		query.Set("pagelen", pageLen)
	}
	if page != "" {
		query.Set("page", page)
	}

	if next != "" {
		overrideURL, err := url.Parse(next)
		if err != nil {
			err = fmt.Errorf("invalid next page URL %q: %w", next, err)
			otel.IncrementAPICallCounter(time.Now().UTC(), name, err)
			return "", nil, err
		}

		path = strings.TrimPrefix(overrideURL.Path, "/2.0") // [API.httpRequestPrep] adds this automatically.
		query = overrideURL.Query()
	}

	return path, query, nil
}
