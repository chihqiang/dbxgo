package output

import (
	"context"
	"encoding/json"
	"fmt"

	"chihqiang/dbxgo/pkg/structx"
	"chihqiang/dbxgo/types"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

// RocketMQConfig RocketMQ configuration entity
type RocketMQConfig struct {
	// Servers - RocketMQ NameServer address list, e.g., ["127.0.0.1:9876"]
	Servers []string `yaml:"servers" json:"servers" mapstructure:"servers" env:"OUTPUT_ROCKETMQ_SERVERS" envDefault:"127.0.0.1:9876"`
	// Topic - The topic name to send the message
	Topic string `yaml:"topic" json:"topic" mapstructure:"topic" env:"OUTPUT_ROCKETMQ_TOPIC" envDefault:"dbxgo-events"`
	// Group - The producer group name
	Group string `yaml:"group" json:"group" mapstructure:"group" env:"OUTPUT_ROCKETMQ_GROUP"`
	// Retry - The number of retries if sending a message fails
	Retry int `yaml:"retry" json:"retry" mapstructure:"retry" env:"OUTPUT_ROCKETMQ_RETRY" envDefault:"3"`
	// Namespace - The namespace
	Namespace string `yaml:"namespace" json:"namespace" mapstructure:"namespace" env:"OUTPUT_ROCKETMQ_NAMESPACE"`
	// AccessKey - Access key
	AccessKey string `yaml:"access_key" json:"access_key" mapstructure:"access_key" env:"OUTPUT_ROCKETMQ_ACCESS_KEY"`
	// SecretKey - Secret key
	SecretKey string `yaml:"secret_key" json:"secret_key" mapstructure:"secret_key" env:"OUTPUT_ROCKETMQ_SECRET_KEY"`
}

// RocketMQOutput RocketMQ implementation that satisfies the IOutput interface
type RocketMQOutput struct {
	cfg      RocketMQConfig
	producer rocketmq.Producer
}

// NewRocketMQOutput Creates a RocketMQOutput and fills in default values
func NewRocketMQOutput(cfg RocketMQConfig) (*RocketMQOutput, error) {
	var (
		err error
	)
	cfg, err = structx.MergeWithDefaults[RocketMQConfig](cfg)
	if err != nil {
		return nil, err
	}
	// Create producer options
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
	// Create producer
	p, err := rocketmq.NewProducer(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create RocketMQ producer: %w", err)
	}
	// Start the producer
	if err := p.Start(); err != nil {
		return nil, fmt.Errorf("failed to start RocketMQ producer: %w", err)
	}
	return &RocketMQOutput{
		cfg:      cfg,
		producer: p,
	}, nil
}

// Send Serializes the EventData to a JSON string and sends it to RocketMQ
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

// Close Closes the RocketMQ producer
func (r *RocketMQOutput) Close() error {
	return r.producer.Shutdown()
}
