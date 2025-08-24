package temporal

import (
	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"
	"go.temporal.io/sdk/client"
)

const (
	DefaultTaskQueue = "timpani"
)

// Flags defines CLI flags to configure a Temporal worker. Usually these flags
// are set using environment variables or the application's configuration file.
func Flags(configFilePath altsrc.StringSourcer) []cli.Flag {
	return []cli.Flag{
		// https://pkg.go.dev/go.temporal.io/sdk/internal#ClientOptions
		&cli.StringFlag{
			Name:  "temporal-address",
			Usage: "Temporal server address",
			Value: client.DefaultHostPort,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("TEMPORAL_ADDRESS"),
				toml.TOML("temporal.address", configFilePath),
			),
		},
		&cli.StringFlag{
			Name:  "temporal-namespace",
			Usage: "Temporal namespace",
			Value: client.DefaultNamespace,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("TEMPORAL_NAMESPACE"),
				toml.TOML("temporal.namespace", configFilePath),
			),
		},

		// Worker parameter.
		&cli.StringFlag{
			Name:  "temporal-task-queue",
			Usage: "Temporal task queue",
			Value: DefaultTaskQueue,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("TEMPORAL_TASK_QUEUE"),
				toml.TOML("temporal.task_queue", configFilePath),
			),
		},

		// https://pkg.go.dev/go.temporal.io/sdk/internal#WorkerOptions
	}
}
