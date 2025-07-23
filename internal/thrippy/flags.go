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

// Flags defines CLI flags to configure a Thrippy gRPC client. These flags can also
// be set using environment variables and the application's configuration file.
func Flags(configFilePath altsrc.StringSourcer) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "thrippy-server-addr",
			Usage: "Thrippy gRPC server address",
			Value: DefaultGRPCAddress,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("THRIPPY_SERVER_ADDRESS"),
				toml.TOML("thrippy.server_address", configFilePath),
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

// LinkIDFlag defines a CLI flag for the Thrippy link ID of a given
// third-party service provider. This flag can also be set using an
// environment variable and the application's configuration file.
func LinkIDFlag(configFilePath altsrc.StringSourcer, sp string) cli.Flag {
	lowerCase := strings.ToLower(sp)
	return &cli.StringFlag{
		Name:  "thrippy-link-" + lowerCase,
		Usage: "Thrippy link ID for " + sp,
		Sources: cli.NewValueSourceChain(
			cli.EnvVar("THRIPPY_LINK_"+strings.ToUpper(sp)),
			toml.TOML("thrippy.links."+lowerCase, configFilePath),
		),
	}
}

// LinkID extracts and checks the configured Thrippy
// link ID for the given third-party service provider.
func LinkID(l zerolog.Logger, cmd *cli.Command, sp string) (string, bool) {
	id := cmd.String("thrippy-link-" + strings.ToLower(sp))
	if id == "" {
		l.Warn().Msg("Thrippy link ID not configured for " + sp)
		return "", false
	}

	if _, err := shortuuid.DefaultEncoder.Decode(id); err != nil {
		l.Error().Msg("invalid Thrippy link ID configured for " + sp)
		return "", false
	}

	return id, true
}
