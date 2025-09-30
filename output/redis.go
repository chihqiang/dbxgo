package output

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chihqiang/dbxgo/types"
	"github.com/redis/go-redis/v9"
)

// RedisConfig Redis 配置实体
type RedisConfig struct {
	// Redis 地址，例如 "127.0.0.1:6379"
	Addr string `json:"addr" yaml:"addr"`
	// Redis 密码，可为空
	Password string `json:"password" yaml:"password"`
	// Redis 数据库编号
	DB int `json:"db" yaml:"db"`
	// 用于存储事件的 key（List名称）
	Key string `json:"key" yaml:"key"`
}

// DefaultRedisConfig 返回 Redis 默认配置
func DefaultRedisConfig() RedisConfig {
	return RedisConfig{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
		Key:      "dbxgo_events",
	}
}

type RedisOutput struct {
	cfg RedisConfig
	rdb *redis.Client
	ctx context.Context
	key string
}

// NewRedisOutput 创建 RedisOutput，并填充默认值
func NewRedisOutput(cfg RedisConfig) (*RedisOutput, error) {
	// 获取默认配置
	def := DefaultRedisConfig()

	// 对每个字段判断零值，如果未设置则使用默认值
	if cfg.Addr == "" {
		cfg.Addr = def.Addr
	}
	if cfg.Password == "" {
		cfg.Password = def.Password
	}
	if cfg.DB == 0 {
		cfg.DB = def.DB
	}
	if cfg.Key == "" {
		cfg.Key = def.Key
	}
	// 创建 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	ctx := context.Background()
	// 测试连接
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect Redis: %w", err)
	}
	return &RedisOutput{
		cfg: cfg,
		rdb: rdb,
		ctx: ctx,
		key: cfg.Key,
	}, nil
}

// Send 将事件发送到 Redis（使用 List）
func (r *RedisOutput) Send(event types.EventData) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	// LPUSH 推送到列表头
	if err := r.rdb.LPush(r.ctx, r.key, data).Err(); err != nil {
		return fmt.Errorf("failed to push event to Redis: %w", err)
	}
	return nil
}

// Close 关闭 Redis 客户端
func (r *RedisOutput) Close() error {
	return r.rdb.Close()
}
