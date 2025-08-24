package main

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli/v3"

	"github.com/tzrikka/timpani/internal/thrippy"
	"github.com/tzrikka/timpani/pkg/http/webhooks"
	"github.com/tzrikka/timpani/pkg/temporal"
	"github.com/tzrikka/xdg"
)

const (
	ConfigDirName  = "timpani"
	ConfigFileName = "config.toml"
)

var services = []string{
	"Bitbucket",
	"GitHub",
	"Slack",
}

func main() {
	bi, _ := debug.ReadBuildInfo()

	cmd := &cli.Command{
		Name:    "timpani",
		Usage:   "Temporal worker that sends API calls and receives event notifications",
		Version: bi.Main.Version,
		Flags:   flags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			initLog(cmd.Bool("dev") || cmd.Bool("pretty-log"))
			s := webhooks.NewHTTPServer(cmd)
			go s.Run()
			if err := s.ConnectLinks(ctx, cmd); err != nil {
				return err
			}
			return temporal.Run(log.Logger, cmd)
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func flags() []cli.Flag {
	fs := []cli.Flag{
		&cli.BoolFlag{
			Name:  "dev",
			Usage: "simple setup, but unsafe for production",
		},
		&cli.BoolFlag{
			Name:  "pretty-log",
			Usage: "human-readable console logging, instead of JSON",
		},
	}

	path := configFile()
	fs = append(fs, temporal.Flags(path)...)
	fs = append(fs, thrippy.Flags(path)...)
	fs = append(fs, webhooks.Flags(path)...)

	for _, s := range services {
		fs = append(fs, thrippy.LinkIDFlag(path, s))
	}

	return fs
}

// configFile returns the path to the app's configuration file.
// It also creates an empty file if it doesn't already exist.
func configFile() altsrc.StringSourcer {
	path, err := xdg.CreateFile(xdg.ConfigHome, ConfigDirName, ConfigFileName)
	if err != nil {
		log.Fatal().Err(err).Caller().Send()
	}
	return altsrc.StringSourcer(path)
}

// initLog initializes the logger for Timpani's HTTP server and Temporal
// worker, based on whether it's running in development mode or not.
func initLog(devMode bool) {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs

	if !devMode {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = zerolog.New(os.Stderr).With().Timestamp().Caller().Logger()
		return
	}

	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05.000",
	}).With().Caller().Logger()

	log.Warn().Msg("********** DEV MODE - UNSAFE IN PRODUCTION! **********")
}
