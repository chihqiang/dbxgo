package redisx

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

// Config Redis configuration struct
type Config struct {
	Addr     string
	Password string
	DB       int
}

// Open Creates a Redis client using the configuration struct
func Open(cfg Config) (*redis.Client, error) {
	// Set default values
	if cfg.Addr == "" {
		cfg.Addr = "127.0.0.1:6379"
	}
	opt := &redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	}
	rdb := redis.NewClient(opt)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect Redis: %w", err)
	}
	return rdb, nil
}
