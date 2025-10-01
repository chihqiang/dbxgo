package store

import (
	"context"
	"fmt"
	"github.com/chihqiang/dbxgo/pkg/redisx"
	"github.com/redis/go-redis/v9"
)

// RedisStore Redis
type RedisStore struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisStore 创建 RedisStore
func NewRedisStore(cfg redisx.Config) (*RedisStore, error) {
	rdb, err := redisx.Open(cfg)
	if err != nil {
		return nil, err
	}
	return &RedisStore{
		ctx:    context.Background(),
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
