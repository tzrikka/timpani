package github

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/tzrikka/timpani-api/pkg/github"
	"github.com/tzrikka/timpani/pkg/metrics"
)

// PullRequestsGetActivity is based on:
// https://docs.github.com/en/rest/pulls/pulls?apiVersion=2022-11-28#get-a-pull-request
func (a *API) PullRequestsGetActivity(ctx context.Context, req github.PullRequestsGetRequest) (*github.PullRequest, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d", req.Owner, req.Repo, req.PullNumber)

	t := time.Now().UTC()
	resp := new(github.PullRequest)
	_, err := a.httpGet(ctx, req.ThrippyLinkID, path, nil, resp)
	metrics.IncrementAPICallCounter(t, github.PullRequestsGetActivityName, err)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

// PullRequestsListCommitsActivity is based on:
// https://docs.github.com/en/rest/pulls/pulls?apiVersion=2022-11-28#list-commits-on-a-pull-request
//
// Pagination is handled internally if both PerPage and Page are 0 in the request, but either way
// the results are limited to a maximum of 250 commits. To receive a complete list, call [CommitsList].
func (a *API) PullRequestsListCommitsActivity(ctx context.Context, req github.PullRequestsListCommitsRequest) ([]github.Commit, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d/commits", req.Owner, req.Repo, req.PullNumber)
	activityName := github.PullRequestsListCommitsActivityName

	return paginatedActivity[github.Commit](ctx, a, activityName, req.ThrippyLinkID, path, req.PerPage, req.Page)
}

// PullRequestsListFilesActivity is based on:
// https://docs.github.com/en/rest/pulls/pulls?apiVersion=2022-11-28#list-pull-requests-files
//
// Pagination is handled internally if both PerPage and Page are 0 in the
// request, but either way the results are limited to a maximum of 3000 files.
func (a *API) PullRequestsListFilesActivity(ctx context.Context, req github.PullRequestsListFilesRequest) ([]github.File, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d/files", req.Owner, req.Repo, req.PullNumber)
	activityName := github.PullRequestsListFilesActivityName

	return paginatedActivity[github.File](ctx, a, activityName, req.ThrippyLinkID, path, req.PerPage, req.Page)
}

func paginatedActivity[T any](ctx context.Context, a *API, activityName, linkID, path string, perPage, page int) ([]T, error) {
	paginate := perPage == 0 && page == 0
	if paginate {
		perPage = 100 // Default = 30, but we prefer to minimize the number of API calls.
		page = 1
	}

	query := url.Values{}
	if perPage != 0 {
		query.Set("per_page", strconv.Itoa(perPage))
	}
	if page != 0 {
		query.Set("page", strconv.Itoa(page))
	}

	var results []T
	hasMore := true

	for hasMore {
		t := time.Now().UTC()
		resp := new([]T)
		more, err := a.httpGet(ctx, linkID, path, query, resp)
		metrics.IncrementAPICallCounter(t, activityName, err)
		if err != nil {
			return nil, err
		}

		results = append(results, *resp...)
		hasMore = paginate && more

		page++
		query.Set("page", strconv.Itoa(page))
	}

	return results, nil
}

// PullRequestsMergeActivity is based on:
// https://docs.github.com/en/rest/pulls/pulls?apiVersion=2022-11-28#merge-a-pull-request
func (a *API) PullRequestsMergeActivity(ctx context.Context, req github.PullRequestsMergeRequest) (*github.PullRequestsMergeResponse, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d/merge", req.Owner, req.Repo, req.PullNumber)

	t := time.Now().UTC()
	resp := new(github.PullRequestsMergeResponse)
	err := a.httpPut(ctx, req.ThrippyLinkID, path, defaultAccept, req, resp)
	metrics.IncrementAPICallCounter(t, github.PullRequestsMergeActivityName, err)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

// PullRequestsCommentsCreateActivity is based on:
// https://docs.github.com/en/rest/pulls/comments?apiVersion=2022-11-28#create-a-review-comment-for-a-pull-request
func (a *API) PullRequestsCommentsCreateActivity(ctx context.Context, req github.PullRequestsCommentsCreateRequest) (*github.PullComment, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d/comments", req.Owner, req.Repo, req.PullNumber)

	linkID := req.ThrippyLinkID
	req.ThrippyLinkID = ""
	req.Owner = ""
	req.Repo = ""
	req.PullNumber = 0

	t := time.Now().UTC()
	resp := new(github.PullComment)
	err := a.httpPost(ctx, linkID, path, defaultAccept, req, resp)
	metrics.IncrementAPICallCounter(t, github.PullRequestsCommentsCreateActivityName, err)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

// PullRequestsCommentsCreateReplyActivity is based on:
// https://docs.github.com/en/rest/pulls/comments?apiVersion=2022-11-28#create-a-reply-for-a-review-comment
func (a *API) PullRequestsCommentsCreateReplyActivity(
	ctx context.Context,
	req github.PullRequestsCommentsCreateReplyRequest,
) (*github.PullComment, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d/comments/%d/replies", req.Owner, req.Repo, req.PullNumber, req.CommentID)

	linkID := req.ThrippyLinkID
	req.ThrippyLinkID = ""
	req.Owner = ""
	req.Repo = ""
	req.PullNumber = 0
	req.CommentID = 0

	t := time.Now().UTC()
	resp := new(github.PullComment)
	err := a.httpPost(ctx, linkID, path, defaultAccept, req, resp)
	metrics.IncrementAPICallCounter(t, github.PullRequestsCommentsCreateReplyActivityName, err)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

// PullRequestsCommentsDeleteActivity is based on:
// https://docs.github.com/en/rest/pulls/comments?apiVersion=2022-11-28#delete-a-review-comment-for-a-pull-request
func (a *API) PullRequestsCommentsDeleteActivity(ctx context.Context, req github.PullRequestsCommentsDeleteRequest) error {
	path := fmt.Sprintf("/repos/%s/%s/pulls/comments/%d", req.Owner, req.Repo, req.CommentID)

	t := time.Now().UTC()
	err := a.httpDelete(ctx, req.ThrippyLinkID, path, nil)
	metrics.IncrementAPICallCounter(t, github.PullRequestsCommentsDeleteActivityName, err)

	if err != nil {
		return err
	}
	return nil
}

// PullRequestsCommentsUpdateActivity is based on:
// https://docs.github.com/en/rest/pulls/comments?apiVersion=2022-11-28#update-a-review-comment-for-a-pull-request
func (a *API) PullRequestsCommentsUpdateActivity(ctx context.Context, req github.PullRequestsCommentsUpdateRequest) (*github.PullComment, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls/comments/%d", req.Owner, req.Repo, req.CommentID)

	linkID := req.ThrippyLinkID
	req.ThrippyLinkID = ""
	req.Owner = ""
	req.Repo = ""
	req.CommentID = 0

	t := time.Now().UTC()
	resp := new(github.PullComment)
	err := a.httpPatch(ctx, linkID, path, "application/vnd.github-commitcomment.raw+json", req, resp)
	metrics.IncrementAPICallCounter(t, github.PullRequestsCommentsUpdateActivityName, err)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

// PullRequestsReviewsCreateActivity is based on:
// https://docs.github.com/en/rest/pulls/reviews?apiVersion=2022-11-28#create-a-review-for-a-pull-request
func (a *API) PullRequestsReviewsCreateActivity(ctx context.Context, req github.PullRequestsReviewsCreateRequest) (*github.Review, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d/reviews", req.Owner, req.Repo, req.PullNumber)

	linkID := req.ThrippyLinkID
	req.ThrippyLinkID = ""
	req.Owner = ""
	req.Repo = ""
	req.PullNumber = 0

	t := time.Now().UTC()
	resp := new(github.Review)
	err := a.httpPost(ctx, linkID, path, "application/vnd.github-commitcomment.raw+json", req, resp)
	metrics.IncrementAPICallCounter(t, github.PullRequestsReviewsCreateActivityName, err)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

// PullRequestsReviewsDeleteActivity is based on:
// https://docs.github.com/en/rest/pulls/reviews?apiVersion=2022-11-28#delete-a-pending-review-for-a-pull-request
func (a *API) PullRequestsReviewsDeleteActivity(ctx context.Context, req github.PullRequestsReviewsDeleteRequest) error {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d/reviews/%d", req.Owner, req.Repo, req.PullNumber, req.ReviewID)

	t := time.Now().UTC()
	err := a.httpDelete(ctx, req.ThrippyLinkID, path, nil)
	metrics.IncrementAPICallCounter(t, github.PullRequestsReviewsDeleteActivityName, err)

	if err != nil {
		return err
	}
	return nil
}

// PullRequestsReviewsDismissActivity is based on:
// https://docs.github.com/en/rest/pulls/reviews?apiVersion=2022-11-28#dismiss-a-review-for-a-pull-request
func (a *API) PullRequestsReviewsDismissActivity(ctx context.Context, req github.PullRequestsReviewsDismissRequest) (*github.Review, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d/reviews/%d/dismissals", req.Owner, req.Repo, req.PullNumber, req.ReviewID)

	linkID := req.ThrippyLinkID
	req.ThrippyLinkID = ""
	req.Owner = ""
	req.Repo = ""
	req.PullNumber = 0
	req.ReviewID = 0

	t := time.Now().UTC()
	resp := new(github.Review)
	err := a.httpPut(ctx, linkID, path, "application/vnd.github-commitcomment.raw+json", req, resp)
	metrics.IncrementAPICallCounter(t, github.PullRequestsReviewsDismissActivityName, err)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

// PullRequestsReviewsSubmitActivity is based on:
// https://docs.github.com/en/rest/pulls/reviews?apiVersion=2022-11-28#submit-a-review-for-a-pull-request
func (a *API) PullRequestsReviewsSubmitActivity(ctx context.Context, req github.PullRequestsReviewsSubmitRequest) (*github.Review, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d/reviews/%d/events", req.Owner, req.Repo, req.PullNumber, req.ReviewID)

	linkID := req.ThrippyLinkID
	req.ThrippyLinkID = ""
	req.Owner = ""
	req.Repo = ""
	req.PullNumber = 0
	req.ReviewID = 0

	t := time.Now().UTC()
	resp := new(github.Review)
	err := a.httpPost(ctx, linkID, path, "application/vnd.github-commitcomment.raw+json", req, resp)
	metrics.IncrementAPICallCounter(t, github.PullRequestsReviewsSubmitActivityName, err)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

// PullRequestsReviewsUpdateActivity is based on:
// https://docs.github.com/en/rest/pulls/reviews?apiVersion=2022-11-28#update-a-review-for-a-pull-request
func (a *API) PullRequestsReviewsUpdateActivity(ctx context.Context, req github.PullRequestsReviewsUpdateRequest) (*github.Review, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d/reviews/%d", req.Owner, req.Repo, req.PullNumber, req.ReviewID)

	linkID := req.ThrippyLinkID
	req.ThrippyLinkID = ""
	req.Owner = ""
	req.Repo = ""
	req.PullNumber = 0
	req.ReviewID = 0

	t := time.Now().UTC()
	resp := new(github.Review)
	err := a.httpPut(ctx, linkID, path, "application/vnd.github-commitcomment.raw+json", req, resp)
	metrics.IncrementAPICallCounter(t, github.PullRequestsReviewsUpdateActivityName, err)

	if err != nil {
		return nil, err
	}
	return resp, nil
}
