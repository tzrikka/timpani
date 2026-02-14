package bitbucket

import (
	"context"

	"github.com/urfave/cli/v3"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"

	"github.com/tzrikka/timpani-api/pkg/bitbucket"
	"github.com/tzrikka/timpani/internal/thrippy"
)

type API struct {
	thrippy thrippy.LinkClient
}

// Register exposes Temporal activities and workflows via the Timpani worker.
func Register(ctx context.Context, cmd *cli.Command, w worker.Worker) {
	id, ok := thrippy.LinkID(cmd, "Bitbucket")
	if !ok {
		return
	}

	a := API{thrippy: thrippy.NewLinkClient(ctx, id, cmd)}

	registerActivity(w, a.CommitsDiffActivity, bitbucket.CommitsDiffActivityName)
	registerActivity(w, a.CommitsDiffstatActivity, bitbucket.CommitsDiffstatActivityName)

	registerActivity(w, a.PullRequestsApproveActivity, bitbucket.PullRequestsApproveActivityName)
	registerActivity(w, a.PullRequestsCreateCommentActivity, bitbucket.PullRequestsCreateCommentActivityName)
	registerActivity(w, a.PullRequestsDeclineActivity, bitbucket.PullRequestsDeclineActivityName)
	registerActivity(w, a.PullRequestsDeleteCommentActivity, bitbucket.PullRequestsDeleteCommentActivityName)
	registerActivity(w, a.PullRequestsDiffstatActivity, bitbucket.PullRequestsDiffstatActivityName)
	registerActivity(w, a.PullRequestsGetActivity, bitbucket.PullRequestsGetActivityName)
	registerActivity(w, a.PullRequestsGetCommentActivity, bitbucket.PullRequestsGetCommentActivityName)
	registerActivity(w, a.PullRequestsListActivityLogActivity, bitbucket.PullRequestsListActivityLogActivityName)
	registerActivity(w, a.PullRequestsListCommitsActivity, bitbucket.PullRequestsListCommitsActivityName)
	registerActivity(w, a.PullRequestsListForCommitActivity, bitbucket.PullRequestsListForCommitActivityName)
	registerActivity(w, a.PullRequestsListTasksActivity, bitbucket.PullRequestsListTasksActivityName)
	registerActivity(w, a.PullRequestsMergeActivity, bitbucket.PullRequestsMergeActivityName)
	registerActivity(w, a.PullRequestsUnapproveActivity, bitbucket.PullRequestsUnapproveActivityName)
	registerActivity(w, a.PullRequestsUpdateActivity, bitbucket.PullRequestsUpdateActivityName)
	registerActivity(w, a.PullRequestsUpdateCommentActivity, bitbucket.PullRequestsUpdateCommentActivityName)

	registerActivity(w, a.SourceGetFileActivity, bitbucket.SourceGetFileActivityName)

	registerActivity(w, a.UsersGetActivity, bitbucket.UsersGetActivityName)

	registerActivity(w, a.WorkspacesListMembersActivity, bitbucket.WorkspacesListMembersActivityName)
}

func registerActivity(w worker.Worker, f any, name string) {
	w.RegisterActivityWithOptions(f, activity.RegisterOptions{Name: name})
}
