package main

import (
	"chihqiang/dbxgo/cmd"
	"chihqiang/dbxgo/pkg/logx"
	"context"
	"fmt"
	"github.com/urfave/cli/v3"
	"os"
	"runtime"
)

var (
	version = "main"
)

func init() {
}

func main() {
	app := &cli.Command{}
	app.Name = "dbxgo"
	app.Usage = "a Go CDC tool that real-time captures, processes database changes and sends them to downstream"
	app.Version = version
	cli.VersionPrinter = func(cmd *cli.Command) {
		_, _ = fmt.Fprintf(cmd.Root().Writer, "%s %s — built with %s on %s/%s\n",
			cmd.Name, cmd.Version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	}
	app.Flags = cmd.Flags()
	app.Before = cmd.Before
	app.Commands = []*cli.Command{
		cmd.ListenCommand(),
	}
	app.Action = func(ctx context.Context, command *cli.Command) error {
		listenCmd := cmd.ListenCommand()
		return listenCmd.Action(ctx, command)
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		logx.Error("%v", err)
		os.Exit(1)
	}
}
