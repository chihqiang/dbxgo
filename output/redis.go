package output

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chihqiang/dbxgo/pkg/redisx"
	"github.com/chihqiang/dbxgo/pkg/x"
	"github.com/chihqiang/dbxgo/types"
	"github.com/redis/go-redis/v9"
)

// RedisConfig Redis 配置实体
type RedisConfig struct {
	Addr     string `yaml:"addr" json:"addr" mapstructure:"addr" env:"OUTPUT_REDIS_ADDR" envDefault:"127.0.0.1:6379"`
	Password string `yaml:"password" json:"password" mapstructure:"password" env:"OUTPUT_REDIS_PASSWORD" envDefault:""`
	DB       int    `yaml:"db" json:"db" mapstructure:"db" env:"OUTPUT_REDIS_DB" envDefault:"0"`
	Key      string `yaml:"key" json:"key" mapstructure:"key" env:"OUTPUT_REDIS_KEY" envDefault:"dbxgo-events"`
}

type RedisOutput struct {
	cfg RedisConfig
	rdb *redis.Client
	key string
}

// NewRedisOutput 创建 RedisOutput，并填充默认值
func NewRedisOutput(cfg RedisConfig) (*RedisOutput, error) {
	var err error
	cfg, err = x.MergeWithDefaults[RedisConfig](cfg)
	if err != nil {
		return nil, err
	}
	rdb, err := redisx.Open(redisx.Config{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err != nil {
		return nil, err
	}
	return &RedisOutput{
		cfg: cfg,
		rdb: rdb,
		key: cfg.Key,
	}, nil
}

// Send 将事件发送到 Redis（使用 List）
func (r *RedisOutput) Send(ctx context.Context, event types.EventData) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	// LPUSH 推送到列表头
	if err := r.rdb.LPush(ctx, r.key, data).Err(); err != nil {
		return fmt.Errorf("failed to push event to Redis: %w", err)
	}
	return nil
}

// Close 关闭 Redis 客户端
func (r *RedisOutput) Close() error {
	return r.rdb.Close()
}
