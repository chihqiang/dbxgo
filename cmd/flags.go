package cmd

import "github.com/urfave/cli/v3"

const (
	FlagConfig = "config"
)

func Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    FlagConfig,
			Aliases: []string{"c"},
			Usage:   "Load configuration from `FILE`",
			Value:   "config.yml",
		},
	}
}
