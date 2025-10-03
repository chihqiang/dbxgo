package redisx

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

// Config Redis 配置结构体
type Config struct {
	Addr     string
	Password string
	DB       int
}

// Open 使用配置结构体创建 Redis 客户端
func Open(cfg Config) (*redis.Client, error) {
	// 设置默认值
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
