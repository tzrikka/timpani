package jira

import (
	"context"

	"github.com/urfave/cli/v3"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"

	"github.com/tzrikka/timpani-api/pkg/jira"
	"github.com/tzrikka/timpani/internal/thrippy"
)

type API struct {
	thrippy thrippy.LinkClient
}

// Register exposes Temporal activities and workflows via the Timpani worker.
func Register(ctx context.Context, cmd *cli.Command, w worker.Worker) {
	id, ok := thrippy.LinkID(cmd, "Jira")
	if !ok {
		return
	}

	a := API{thrippy: thrippy.NewLinkClient(ctx, id, cmd)}

	registerActivity(w, a.UsersGetActivity, jira.UsersGetActivityName)
	registerActivity(w, a.UsersSearchActivity, jira.UsersSearchActivityName)
}

func registerActivity(w worker.Worker, f any, name string) {
	w.RegisterActivityWithOptions(f, activity.RegisterOptions{Name: name})
}
