package main

import (
	"context"
	"fmt"
	"github.com/chihqiang/dbxgo/cmd"
	"github.com/urfave/cli/v3"
	"log"
	"log/slog"
	"os"
	"runtime"
)

var (
	version = "main"
)

func init() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
}

func main() {
	app := &cli.Command{}
	app.Name = "dbxgo"
	app.Usage = "a Go CDC tool that real-time captures, processes database changes and sends them to downstream"
	app.Version = version
	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Printf("dbxgo version %s %s/%s\n", cmd.Version, runtime.GOOS, runtime.GOARCH)
	}
	app.Flags = cmd.Flags()
	app.Before = cmd.Before
	app.Commands = []*cli.Command{
		cmd.ListenCommand(),
	}
	app.Action = func(ctx context.Context, command *cli.Command) error {
		startCmd := cmd.ListenCommand()
		return startCmd.Action(ctx, startCmd)
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
