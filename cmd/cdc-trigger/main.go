package main

import (
	"context"
	"os"
	"runtime"

	"github.com/chihqiang/cdc-trigger/cmd"
	"github.com/chihqiang/cli"
	"github.com/chihqiang/logx"
)

var (
	version = "main"
)

func init() {
}

func main() {
	app := &cli.Command{}
	app.Name = "cdc-trigger"
	app.Usage = "a Go CDC tool that real-time captures, processes database changes and sends them to downstream"
	app.Version = version
	app.VersionPrinter = func(ctx context.Context, in *cli.Input, out *cli.Output) {
		out.Infof("%s %s — built with %s on %s/%s",
			in.Command().Name, in.Command().Version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	}
	app.Flags = cmd.Flags()
	app.Before = cmd.Before
	app.Subcommands = []*cli.Command{
		cmd.ListenCommand(),
	}
	app.Action = func(ctx context.Context, in *cli.Input, out *cli.Output) error {
		listenCmd := cmd.ListenCommand()
		return listenCmd.Action(ctx, in, out)
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		logx.Error("%v", err)
		os.Exit(1)
	}
}
