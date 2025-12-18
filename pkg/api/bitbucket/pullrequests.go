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
			metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsCreateCommentActivityName, err)
			return nil, temporal.NewNonRetryableApplicationError("invalid parent ID", fmt.Sprintf("%T", err), err, req.ParentID)
		}
		body.Parent = &prCreateCommentParent{ID: id}
	}

	resp := new(bitbucket.PullRequestsCreateCommentResponse)
	err := a.httpPost(ctx, req.ThrippyLinkID, path, body, resp)
	metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsCreateCommentActivityName, err)

	if err != nil {
		return nil, err
	}
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

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-diffstat-get
func (a *API) PullRequestsDiffstatActivity(
	ctx context.Context,
	req bitbucket.PullRequestsDiffstatRequest,
) (*bitbucket.PullRequestsDiffstatResponse, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/diffstat", req.Workspace, req.RepoSlug, req.PullRequestID)
	query := paginatedQuery(req.PageLen, req.Page)

	t := time.Now().UTC()
	if req.Next != "" {
		overrideURL, err := url.Parse(req.Next)
		if err != nil {
			err = fmt.Errorf("invalid next page URL %q: %w", req.Next, err)
			metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsDiffstatActivityName, err)
			return nil, err
		}

		path = strings.TrimPrefix(overrideURL.Path, "/2.0") // [API.httpRequestPrep] adds this automatically.
		query = overrideURL.Query()
	}

	resp := new(bitbucket.PullRequestsDiffstatResponse)
	err := a.httpGet(ctx, req.ThrippyLinkID, path, query, resp)
	metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsDiffstatActivityName, err)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-get
func (a *API) PullRequestsGetActivity(ctx context.Context, req bitbucket.PullRequestsGetRequest) (map[string]any, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s", req.Workspace, req.RepoSlug, req.PullRequestID)

	t := time.Now().UTC()
	resp := map[string]any{}
	err := a.httpGet(ctx, req.ThrippyLinkID, path, url.Values{}, &resp)
	metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsGetActivityName, err)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-activity-get
func (a *API) PullRequestsListActivityLogActivity(
	ctx context.Context,
	req bitbucket.PullRequestsListActivityLogRequest,
) (*bitbucket.PullRequestsListActivityLogResponse, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/activity", req.Workspace, req.RepoSlug, req.PullRequestID)
	query := paginatedQuery(req.PageLen, req.Page)

	t := time.Now().UTC()
	if req.Next != "" {
		overrideURL, err := url.Parse(req.Next)
		if err != nil {
			err = fmt.Errorf("invalid next page URL %q: %w", req.Next, err)
			metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsListActivityLogActivityName, err)
			return nil, err
		}

		path = strings.TrimPrefix(overrideURL.Path, "/2.0") // [API.httpRequestPrep] adds this automatically.
		query = overrideURL.Query()
	}

	resp := new(bitbucket.PullRequestsListActivityLogResponse)
	err := a.httpGet(ctx, req.ThrippyLinkID, path, query, resp)
	metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsListActivityLogActivityName, err)

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
	query := paginatedQuery(req.PageLen, req.Page)

	t := time.Now().UTC()
	if req.Next != "" {
		overrideURL, err := url.Parse(req.Next)
		if err != nil {
			err = fmt.Errorf("invalid next page URL %q: %w", req.Next, err)
			metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsListCommitsActivityName, err)
			return nil, err
		}

		path = strings.TrimPrefix(overrideURL.Path, "/2.0") // [API.httpRequestPrep] adds this automatically.
		query = overrideURL.Query()
	}

	resp := new(bitbucket.PullRequestsListCommitsResponse)
	err := a.httpGet(ctx, req.ThrippyLinkID, path, query, resp)
	metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsListCommitsActivityName, err)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-commit-commit-pullrequests-get
func (a *API) PullRequestsListForCommitActivity(
	ctx context.Context,
	req bitbucket.PullRequestsListForCommitRequest,
) (*bitbucket.PullRequestsListForCommitResponse, error) {
	path := fmt.Sprintf("/repositories/%s/%s/commit/%s/pullrequests", req.Workspace, req.RepoSlug, req.Commit)
	query := paginatedQuery(req.PageLen, req.Page)

	t := time.Now().UTC()
	if req.Next != "" {
		overrideURL, err := url.Parse(req.Next)
		if err != nil {
			err = fmt.Errorf("invalid next page URL %q: %w", req.Next, err)
			metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsListForCommitActivityName, err)
			return nil, err
		}

		path = strings.TrimPrefix(overrideURL.Path, "/2.0") // [API.httpRequestPrep] adds this automatically.
		query = overrideURL.Query()
	}

	resp := new(bitbucket.PullRequestsListForCommitResponse)
	err := a.httpGet(ctx, req.ThrippyLinkID, path, query, resp)
	metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsListForCommitActivityName, err)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

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
	metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsMergeActivityName, err)

	return err
}

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-approve-delete
func (a *API) PullRequestsUnapproveActivity(ctx context.Context, req bitbucket.PullRequestsUnapproveRequest) error {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/approve", req.Workspace, req.RepoSlug, req.PullRequestID)

	t := time.Now().UTC()
	err := a.httpDelete(ctx, req.ThrippyLinkID, path, url.Values{})
	metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsUnapproveActivityName, err)

	return err
}

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-put
func (a *API) PullRequestsUpdateActivity(ctx context.Context, req bitbucket.PullRequestsUpdateRequest) (map[string]any, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s", req.Workspace, req.RepoSlug, req.PullRequestID)

	t := time.Now().UTC()
	resp := map[string]any{}
	err := a.httpPut(ctx, req.ThrippyLinkID, path, req.PullRequest, &resp)
	metrics.IncrementAPICallCounter(t, bitbucket.PullRequestsUpdateActivityName, err)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-comments-comment-id-put
func (a *API) PullRequestsUpdateCommentActivity(ctx context.Context, req bitbucket.PullRequestsUpdateCommentRequest) error {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%s/comments/%s", req.Workspace, req.RepoSlug, req.PullRequestID, req.CommentID)
	body := &prCommentBody{Content: prCommentContent{Raw: req.Markdown}}

	t := time.Now().UTC()
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
