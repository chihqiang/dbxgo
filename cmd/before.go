package cmd

import (
	"context"
	"chihqiang/dbxgo/config"
	"github.com/urfave/cli/v3"
)

type ContextValue string

var (
	ContextValueConfig ContextValue = "config"
)

// Before loads the configuration file and returns
func Before(ctx context.Context, command *cli.Command) (context.Context, error) {
	conf, err := config.Load(command.String(FlagConfig))
	if err != nil {
		return ctx, err
	}
	//Put the configuration into the context
	return context.WithValue(ctx, ContextValueConfig, conf), nil
}
