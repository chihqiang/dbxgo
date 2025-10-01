package redisx

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	Addr     string `yaml:"addr" json:"addr" mapstructure:"addr" env:"REDIS_ADDR" envDefault:"127.0.0.1:6379"`
	Password string `yaml:"password" json:"password" mapstructure:"password" env:"REDIS_PASSWORD" envDefault:""`
	DB       int    `yaml:"db" json:"db" mapstructure:"db" env:"REDIS_DB" envDefault:"0"`
}

func Open(cfg Config) (*redis.Client, error) {
	if cfg.Addr != "" {
		cfg.Addr = "127.0.0.1:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect Redis: %w", err)
	}
	return rdb, nil
}
