package store

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string `yaml:"addr"` // "127.0.0.1:6379"
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// DefaultRedisConfig 返回 Redis 默认配置
func DefaultRedisConfig() RedisConfig {
	return RedisConfig{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	}
}

// RedisStore Redis
type RedisStore struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisStore 创建 RedisStore
func NewRedisStore(cfg RedisConfig) (*RedisStore, error) {
	def := DefaultRedisConfig()
	if cfg.Addr != "" {
		def.Addr = cfg.Addr
	}
	if cfg.Password != "" {
		def.Password = cfg.Password
	}
	if cfg.DB > 0 {
		def.DB = cfg.DB
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     def.Addr,
		Password: def.Password,
		DB:       def.DB,
	})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect Redis: %w", err)
	}
	return &RedisStore{
		ctx:    ctx,
		client: rdb,
	}, nil
}

// Has 判断 key 是否存在
func (r *RedisStore) Has(key string) bool {
	exists, err := r.client.Exists(r.ctx, key).Result()
	if err != nil {
		return false
	}
	return exists > 0
}

// Set 设置 key 对应的值
func (r *RedisStore) Set(key string, value []byte) error {
	return r.client.Set(r.ctx, key, value, 0).Err()
}

// Get 获取 key 对应的值
func (r *RedisStore) Get(key string) ([]byte, error) {
	val, err := r.client.Get(r.ctx, key).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("key %s does not exist", key)
	}
	return val, err
}
