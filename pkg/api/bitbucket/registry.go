package bitbucket

import (
	"github.com/rs/zerolog"
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
func Register(l zerolog.Logger, cmd *cli.Command, w worker.Worker) {
	id, ok := thrippy.LinkID(l, cmd, "Bitbucket")
	if !ok {
		return
	}

	a := API{thrippy: thrippy.NewLinkClient(id, cmd)}

	registerActivity(w, a.PullRequestsCreateCommentActivity, bitbucket.PullRequestsCreateCommentActivityName)
	registerActivity(w, a.PullRequestsDeleteCommentActivity, bitbucket.PullRequestsDeleteCommentActivityName)
	registerActivity(w, a.PullRequestsListCommitsActivity, bitbucket.PullRequestsListCommitsActivityName)
	registerActivity(w, a.PullRequestsUpdateCommentActivity, bitbucket.PullRequestsUpdateCommentActivityName)

	registerActivity(w, a.UsersGetActivity, bitbucket.UsersGetActivityName)

	registerActivity(w, a.WorkspacesListMembersActivity, bitbucket.WorkspacesListMembersActivityName)
}

func registerActivity(w worker.Worker, f any, name string) {
	w.RegisterActivityWithOptions(f, activity.RegisterOptions{Name: name})
}
