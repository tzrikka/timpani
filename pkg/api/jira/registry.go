package jira

import (
	"github.com/rs/zerolog"
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
func Register(l zerolog.Logger, cmd *cli.Command, w worker.Worker) {
	id, ok := thrippy.LinkID(l, cmd, "Jira")
	if !ok {
		return
	}

	a := API{thrippy: thrippy.NewLinkClient(id, cmd)}

	registerActivity(w, a.UsersGetActivity, jira.UsersGetActivityName)
}

func registerActivity(w worker.Worker, f any, name string) {
	w.RegisterActivityWithOptions(f, activity.RegisterOptions{Name: name})
}
