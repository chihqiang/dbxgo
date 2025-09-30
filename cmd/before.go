package cmd

import (
	"context"
	"github.com/chihqiang/dbxgo/config"
	"github.com/urfave/cli/v3"
)

type CliContextValue string

var (
	CliContextValueConfig CliContextValue = "config"
)

// Before 加载配置文件并返回
func Before(ctx context.Context, command *cli.Command) (context.Context, error) {
	filename := command.String(FlagConfig)
	conf, err := config.Load(filename)
	if err != nil {
		return ctx, err
	}
	// 将配置存储在context中，而不是使用全局变量
	return context.WithValue(ctx, CliContextValueConfig, conf), nil
}
