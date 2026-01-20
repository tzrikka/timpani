package github

import (
	"context"

	"github.com/urfave/cli/v3"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"

	"github.com/tzrikka/timpani-api/pkg/github"
	"github.com/tzrikka/timpani/internal/thrippy"
)

type API struct {
	thrippy thrippy.LinkClient
}

// Register exposes Temporal activities and workflows via the Timpani worker.
func Register(ctx context.Context, cmd *cli.Command, w worker.Worker) {
	id, ok := thrippy.LinkID(cmd, "GitHub")
	if !ok {
		return
	}

	a := API{thrippy: thrippy.NewLinkClient(ctx, id, cmd)}

	registerActivity(w, a.IssuesCommentsCreateActivity, github.IssuesCommentsCreateActivityName)
	registerActivity(w, a.IssuesCommentsDeleteActivity, github.IssuesCommentsDeleteActivityName)
	registerActivity(w, a.IssuesCommentsUpdateActivity, github.IssuesCommentsUpdateActivityName)

	registerActivity(w, a.PullRequestsGetActivity, github.PullRequestsGetActivityName)
	registerActivity(w, a.PullRequestsListCommitsActivity, github.PullRequestsListCommitsActivityName)
	registerActivity(w, a.PullRequestsListFilesActivity, github.PullRequestsListFilesActivityName)
	registerActivity(w, a.PullRequestsMergeActivity, github.PullRequestsMergeActivityName)
	registerActivity(w, a.PullRequestsCommentsCreateActivity, github.PullRequestsCommentsCreateActivityName)
	registerActivity(w, a.PullRequestsCommentsCreateReplyActivity, github.PullRequestsCommentsCreateReplyActivityName)
	registerActivity(w, a.PullRequestsCommentsDeleteActivity, github.PullRequestsCommentsDeleteActivityName)
	registerActivity(w, a.PullRequestsCommentsUpdateActivity, github.PullRequestsCommentsUpdateActivityName)
	registerActivity(w, a.PullRequestsReviewsCreateActivity, github.PullRequestsReviewsCreateActivityName)
	registerActivity(w, a.PullRequestsReviewsDeleteActivity, github.PullRequestsReviewsDeleteActivityName)
	registerActivity(w, a.PullRequestsReviewsDismissActivity, github.PullRequestsReviewsDismissActivityName)
	registerActivity(w, a.PullRequestsReviewsSubmitActivity, github.PullRequestsReviewsSubmitActivityName)
	registerActivity(w, a.PullRequestsReviewsUpdateActivity, github.PullRequestsReviewsUpdateActivityName)

	registerActivity(w, a.UsersGetActivity, github.UsersGetActivityName)
	registerActivity(w, a.UsersListActivity, github.UsersListActivityName)
}

func registerActivity(w worker.Worker, f any, name string) {
	w.RegisterActivityWithOptions(f, activity.RegisterOptions{Name: name})
}
