package thrippy

import (
	"strings"

	"github.com/lithammer/shortuuid/v4"
	"github.com/rs/zerolog"
	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"
)

const (
	DefaultGRPCAddress = "localhost:14460"
)

// Flags defines CLI flags to configure a Thrippy gRPC client. Usually these flags
// are set using environment variables or the application's configuration file.
func Flags(configFilePath altsrc.StringSourcer) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "thrippy-grpc-address",
			Usage: "Thrippy gRPC server address",
			Value: DefaultGRPCAddress,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("THRIPPY_GRPC_ADDRESS"),
				toml.TOML("thrippy.grpc_address", configFilePath),
			),
		},
		&cli.StringFlag{
			Name:  "thrippy-client-cert",
			Usage: "Thrippy gRPC client's public certificate PEM file (mTLS only)",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("THRIPPY_CLIENT_CERT"),
				toml.TOML("thrippy.client_cert", configFilePath),
			),
			TakesFile: true,
		},
		&cli.StringFlag{
			Name:  "thrippy-client-key",
			Usage: "Thrippy gRPC client's private key PEM file (mTLS only)",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("THRIPPY_CLIENT_KEY"),
				toml.TOML("thrippy.client_key", configFilePath),
			),
			TakesFile: true,
		},
		&cli.StringFlag{
			Name:  "thrippy-server-ca-cert",
			Usage: "Thrippy gRPC server's CA certificate PEM file (both TLS and mTLS)",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("THRIPPY_SERVER_CA_CERT"),
				toml.TOML("thrippy.server_ca_cert", configFilePath),
			),
			TakesFile: true,
		},
		&cli.StringFlag{
			Name:  "thrippy-server-name-override",
			Usage: "Thrippy gRPC server's name override (for testing, both TLS and mTLS)",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("THRIPPY_SERVER_NAME_OVERRIDE"),
				toml.TOML("thrippy.server_name_override", configFilePath),
			),
		},
	}
}

// LinkIDFlag defines a CLI flag to specify the Thrippy link ID of a third-party service.
// Usually this flag is set using an environment variable or the application's configuration file.
func LinkIDFlag(configFilePath altsrc.StringSourcer, service string) cli.Flag {
	lowerCase := strings.ToLower(service)
	return &cli.StringFlag{
		Name:  "thrippy-link-" + lowerCase,
		Usage: "Thrippy link ID for " + service,
		Sources: cli.NewValueSourceChain(
			cli.EnvVar("THRIPPY_LINK_"+strings.ToUpper(service)),
			toml.TOML("thrippy.links."+lowerCase, configFilePath),
		),
		Validator: validateOptionalUUID,
	}
}

// LinkID extracts and checks the configured Thrippy link ID for the given third-party service.
func LinkID(l zerolog.Logger, cmd *cli.Command, service string) (string, bool) {
	id := cmd.String("thrippy-link-" + strings.ToLower(service))
	if id == "" {
		l.Warn().Msg("Thrippy link ID not configured for " + service)
		return "", false
	}

	if _, err := shortuuid.DefaultEncoder.Decode(id); err != nil {
		l.Error().Msg("invalid Thrippy link ID configured for " + service)
		return "", false
	}

	return id, true
}

func validateOptionalUUID(id string) error {
	_, err := shortuuid.DefaultEncoder.Decode(id)
	return err
}
