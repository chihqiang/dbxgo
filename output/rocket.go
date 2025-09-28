package output

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/chihqiang/dbxgo/types"
)

// RocketMQConfig RocketMQ 配置实体
type RocketMQConfig struct {
	// Servers 地址列表，例如 ["127.0.0.1:9876"]
	Servers []string `yaml:"servers"`
	// Topic 消息发送到的 topic 名称
	Topic string `yaml:"topic"`
	// Group 消息生产者分组名称
	Group string `yaml:"group"`
	// Retry 发送失败重试次数
	Retry int `yaml:"retry"`
}

// DefaultRocketMQConfig 返回默认 RocketMQ 配置
func DefaultRocketMQConfig() RocketMQConfig {
	return RocketMQConfig{
		Servers: []string{"127.0.0.1:9876"},
		Topic:   "dbxgo",
		Group:   "dbxgo",
		Retry:   3,
	}
}

// RocketMQOutput RocketMQ 实现，满足 IOutput 接口
type RocketMQOutput struct {
	cfg      RocketMQConfig
	producer rocketmq.Producer
	ctx      context.Context
}

// NewRocketMQOutput 创建 RocketMQOutput 并填充默认值
func NewRocketMQOutput(cfg RocketMQConfig) (*RocketMQOutput, error) {
	// 填充默认值
	def := DefaultRocketMQConfig()
	if len(cfg.Servers) == 0 {
		cfg.Servers = def.Servers
	}
	if cfg.Topic == "" {
		cfg.Topic = def.Topic
	}
	if cfg.Group == "" {
		cfg.Group = def.Group
	}

	// 创建生产者配置
	options := []producer.Option{
		producer.WithNameServer(cfg.Servers),
		producer.WithGroupName(cfg.Group),
		producer.WithRetry(cfg.Retry),
	}

	// 创建生产者
	p, err := rocketmq.NewProducer(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create RocketMQ producer: %w", err)
	}
	// 启动生产者
	if err := p.Start(); err != nil {
		return nil, fmt.Errorf("failed to start RocketMQ producer: %w", err)
	}
	return &RocketMQOutput{
		cfg:      cfg,
		producer: p,
		ctx:      context.Background(),
	}, nil
}

// Send 将 EventData 序列化为 JSON 字符串并发送到 RocketMQ
func (r *RocketMQOutput) Send(event types.EventData) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	msg := &primitive.Message{
		Topic: r.cfg.Topic,
		Body:  data,
	}
	_, err = r.producer.SendSync(r.ctx, msg)
	return err
}

// Close 关闭 RocketMQ 生产者
func (r *RocketMQOutput) Close() error {
	return r.producer.Shutdown()
}
