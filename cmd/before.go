package cmd

import (
	"context"
	"github.com/chihqiang/dbxgo/config"
	"github.com/urfave/cli/v3"
)

var (
	cfg *config.Config
)

func Before(ctx context.Context, command *cli.Command) (context.Context, error) {
	filename := command.String(FlagConfig)
	conf, err := config.Load(filename)
	if err != nil {
		return ctx, err
	}
	cfg = conf
	return ctx, nil
}
