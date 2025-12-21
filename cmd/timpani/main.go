package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli/v3"

	"github.com/tzrikka/timpani/internal/logger"
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
	"Jira",
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
			s := webhooks.NewHTTPServer(ctx, cmd)
			go s.Run()
			if err := s.ConnectLinks(ctx); err != nil {
				return err
			}
			return temporal.Run(ctx, cmd)
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
		logger.FatalError("failed to create config file", err)
	}
	return altsrc.StringSourcer(path)
}

// initLog initializes the logger for Timpani's HTTP server and Temporal
// worker, based on whether it's running in development mode or not.
func initLog(devMode bool) {
	var handler slog.Handler
	if devMode {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		})
	}

	slog.SetDefault(slog.New(handler))
}
