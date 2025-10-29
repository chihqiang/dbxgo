package output

import (
	"chihqiang/dbxgo/pkg/redisx"
	"chihqiang/dbxgo/pkg/structx"
	"chihqiang/dbxgo/types"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
)

// RedisConfig Redis configuration entity
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

// NewRedisOutput Creates a RedisOutput and fills in default values
func NewRedisOutput(cfg RedisConfig) (*RedisOutput, error) {
	var err error
	cfg, err = structx.MergeWithDefaults[RedisConfig](cfg)
	if err != nil {
		return nil, err
	}
	// Initialize the Redis client
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

// Send Sends the event to Redis (using List)
func (r *RedisOutput) Send(ctx context.Context, event types.EventData) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	// LPUSH pushes the event to the list head
	if err := r.rdb.LPush(ctx, r.key, data).Err(); err != nil {
		return fmt.Errorf("failed to push event to Redis: %w", err)
	}
	return nil
}

// Close Closes the Redis client
func (r *RedisOutput) Close() error {
	return r.rdb.Close()
}
