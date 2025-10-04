package output

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/chihqiang/dbxgo/pkg/x"
	"github.com/chihqiang/dbxgo/types"
)

// RocketMQConfig RocketMQ 配置实体
type RocketMQConfig struct {
	// Servers - RocketMQ NameServer 地址列表，例如 ["127.0.0.1:9876"]
	Servers []string `yaml:"servers" json:"servers" mapstructure:"servers" env:"OUTPUT_ROCKETMQ_SERVERS" envDefault:"127.0.0.1:9876"`

	// Topic - 消息发送的 Topic 名称
	Topic string `yaml:"topic" json:"topic" mapstructure:"topic" env:"OUTPUT_ROCKETMQ_TOPIC" envDefault:"dbxgo-events"`

	// Group - 生产者分组名称
	Group string `yaml:"group" json:"group" mapstructure:"group" env:"OUTPUT_ROCKETMQ_GROUP"`

	// Namespace - 命名空间
	Namespace string `yaml:"namespace" json:"namespace" mapstructure:"namespace" env:"OUTPUT_ROCKETMQ_NAMESPACE"`

	// AccessKey - 访问密钥 AccessKey
	AccessKey string `yaml:"access_key" json:"access_key" mapstructure:"access_key" env:"OUTPUT_ROCKETMQ_ACCESS_KEY"`

	// SecretKey - 访问密钥 SecretKey
	SecretKey string `yaml:"secret_key" json:"secret_key" mapstructure:"secret_key" env:"OUTPUT_ROCKETMQ_SECRET_KEY"`

	// Retry - 消息发送失败时的重试次数
	Retry int `yaml:"retry" json:"retry" mapstructure:"retry" env:"OUTPUT_ROCKETMQ_RETRY" envDefault:"3"`
}

// RocketMQOutput RocketMQ 实现，满足 IOutput 接口
type RocketMQOutput struct {
	cfg      RocketMQConfig
	producer rocketmq.Producer
}

// NewRocketMQOutput 创建 RocketMQOutput 并填充默认值
func NewRocketMQOutput(cfg RocketMQConfig) (*RocketMQOutput, error) {
	var (
		err error
	)
	cfg, err = x.MergeWithDefaults[RocketMQConfig](cfg)
	if err != nil {
		return nil, err
	}
	// 创建生产者配置
	options := []producer.Option{
		producer.WithNsResolver(primitive.NewPassthroughResolver(cfg.Servers)),
		producer.WithRetry(cfg.Retry),
	}
	if cfg.Group != "" {
		options = append(options, producer.WithGroupName(cfg.Group))
	}
	if cfg.Namespace != "" {
		options = append(options, producer.WithNamespace(cfg.Namespace))
	}
	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		options = append(options, producer.WithCredentials(primitive.Credentials{
			AccessKey: cfg.AccessKey,
			SecretKey: cfg.SecretKey,
		}))
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
	}, nil
}

// Send 将 EventData 序列化为 JSON 字符串并发送到 RocketMQ
func (r *RocketMQOutput) Send(ctx context.Context, event types.EventData) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	msg := &primitive.Message{
		Topic: r.cfg.Topic,
		Body:  data,
	}
	_, err = r.producer.SendSync(ctx, msg)
	return err
}

// Close 关闭 RocketMQ 生产者
func (r *RocketMQOutput) Close() error {
	return r.producer.Shutdown()
}
