package webhooks

import (
	"errors"

	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"
)

const (
	DefaultWebhookPort = 14480
)

// Flags defines CLI flags to configure an HTTP server. Usually these flags
// are set using environment variables or the application's configuration file.
func Flags(configFilePath altsrc.StringSourcer) []cli.Flag {
	return []cli.Flag{
		&cli.IntFlag{
			Name:  "webhook-port",
			Usage: "local port number for HTTP webhooks",
			Value: DefaultWebhookPort,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("TIMPANI_WEBHOOK_PORT"),
				toml.TOML("http_server.webhook_port", configFilePath),
			),
			Validator: validatePort,
		},
		&cli.StringFlag{
			Name:  "thrippy-http-address",
			Usage: "optional Thrippy address, to pass-through OAuth callbacks, to share a single HTTP tunnel",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("THRIPPY_HTTP_PASSTHROUGH_ADDRESS"),
				toml.TOML("http_server.thrippy_http_passthrough_address", configFilePath),
			),
		},
	}
}

func validatePort(p int) error {
	if p < 0 || p > 65535 {
		return errors.New("out of range [0-65535]")
	}
	return nil
}
