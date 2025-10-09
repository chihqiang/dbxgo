package store

import (
	"context"
	"fmt"
	"github.com/chihqiang/dbxgo/pkg/redisx"
	"github.com/chihqiang/dbxgo/pkg/structsx"
	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addr     string `yaml:"addr" json:"addr" mapstructure:"addr" env:"STORE_REDIS_ADDR" envDefault:"127.0.0.1:6379"`
	Password string `yaml:"password" json:"password" mapstructure:"password" env:"STORE_REDIS_PASSWORD" envDefault:""`
	DB       int    `yaml:"db" json:"db" mapstructure:"db" env:"STORE_REDIS_DB" envDefault:"0"`
}

// RedisStore Redis
type RedisStore struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisStore 创建 RedisStore
func NewRedisStore(cfg RedisConfig) (*RedisStore, error) {
	var err error
	cfg, err = structsx.MergeWithDefaults[RedisConfig](cfg)
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
func (r *RedisStore) Delete(key string) error {
	err := r.client.Del(r.ctx, key).Err()
	if err == redis.Nil {
		// key 不存在也算删除成功
		return nil
	}
	return err
}
