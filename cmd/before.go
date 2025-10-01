package cmd

import (
	"context"
	"github.com/chihqiang/dbxgo/config"
	"github.com/urfave/cli/v3"
)

type ContextValue string

var (
	ContextValueConfig ContextValue = "config"
)

// Before 加载配置文件并返回
func Before(ctx context.Context, command *cli.Command) (context.Context, error) {
	conf, err := config.Load(command.String(FlagConfig))
	if err != nil {
		return ctx, err
	}
	// 把配置放到 context 中
	return context.WithValue(ctx, ContextValueConfig, conf), nil
}
