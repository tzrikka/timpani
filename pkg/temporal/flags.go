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

// Flags defines CLI flags to configure a Temporal worker. These flags can also
// be set using environment variables and the application's configuration file.
func Flags(configFilePath altsrc.StringSourcer) []cli.Flag {
	return []cli.Flag{
		// https://pkg.go.dev/go.temporal.io/sdk/internal#ClientOptions
		&cli.StringFlag{
			Name:  "temporal-host-port",
			Usage: "Temporal server address",
			Value: client.DefaultHostPort,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("TEMPORAL_HOST_PORT"),
				toml.TOML("temporal.host_port", configFilePath),
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
