package store

import (
	"context"
	"fmt"
	"chihqiang/dbxgo/pkg/redisx"
	"chihqiang/dbxgo/pkg/structsx"
	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addr     string `yaml:"addr" json:"addr" mapstructure:"addr" env:"STORE_REDIS_ADDR" envDefault:"127.0.0.1:6379"`
	Password string `yaml:"password" json:"password" mapstructure:"password" env:"STORE_REDIS_PASSWORD" envDefault:""`
	DB       int    `yaml:"db" json:"db" mapstructure:"db" env:"STORE_REDIS_DB" envDefault:"0"`
}

// RedisStore Redis store implementation
type RedisStore struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisStore Creates a new RedisStore
func NewRedisStore(cfg RedisConfig) (*RedisStore, error) {
	var err error
	// Merges the configuration with default values
	cfg, err = structsx.MergeWithDefaults[RedisConfig](cfg)
	if err != nil {
		return nil, err
	}
	// Initializes the Redis client
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

// Has Checks if the key exists
func (r *RedisStore) Has(key string) bool {
	exists, err := r.client.Exists(r.ctx, key).Result()
	if err != nil {
		return false
	}
	return exists > 0
}

// Set Sets the value for the given key
func (r *RedisStore) Set(key string, value []byte) error {
	return r.client.Set(r.ctx, key, value, 0).Err()
}

// Get Gets the value of the given key
func (r *RedisStore) Get(key string) ([]byte, error) {
	val, err := r.client.Get(r.ctx, key).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("key %s does not exist", key)
	}
	return val, err
}

// Delete Deletes the given key
func (r *RedisStore) Delete(key string) error {
	err := r.client.Del(r.ctx, key).Err()
	if err == redis.Nil {
		// Consider it successful if the key does not exist
		return nil
	}
	return err
}
func (r *RedisStore) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}
