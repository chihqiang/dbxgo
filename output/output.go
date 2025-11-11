package output

import (
	"chihqiang/dbxgo/types"
	"context"
	"time"
)

type OutputType string

var (
	OutputTypeStdout   OutputType = "stdout"
	OutputTypeRedis    OutputType = "redis"
	OutputTypeKafka    OutputType = "kafka"
	OutputTypeRabbitMQ OutputType = "rabbitmq"
	OutputTypeRocketMQ OutputType = "rocketmq"
	OutputTypePulsar   OutputType = "pulsar"
	outputs                       = map[OutputType]func(Config) (IOutput, error){}
)

func init() {
	Register(OutputTypeStdout, func(config Config) (IOutput, error) {
		return NewStdoutOutput()
	})
	Register(OutputTypeRedis, func(cfg Config) (IOutput, error) {
		return NewRedisOutput(cfg.Redis)
	})
	Register(OutputTypeKafka, func(cfg Config) (IOutput, error) {
		return NewKafkaOutput(cfg.Kafka)
	})
	Register(OutputTypeRabbitMQ, func(cfg Config) (IOutput, error) {
		return NewRabbitMQOutput(cfg.RabbitMQ)
	})
	Register(OutputTypeRocketMQ, func(cfg Config) (IOutput, error) {
		return NewRocketMQOutput(cfg.RocketMQ)
	})
	Register(OutputTypePulsar, func(cfg Config) (IOutput, error) {
		return NewPulsarOutput(cfg.Pulsar)
	})
}

func Register(outputType OutputType, fn func(Config) (IOutput, error)) {
	outputs[outputType] = fn
}

type Config struct {
	Type     OutputType     `yaml:"type" json:"type" mapstructure:"type" env:"OUTPUT_TYPE,required"`
	Redis    RedisConfig    `yaml:"redis" json:"redis" mapstructure:"redis"`
	Kafka    KafkaConfig    `yaml:"kafka" json:"kafka" mapstructure:"kafka"`
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq" json:"rabbitmq" mapstructure:"rabbitmq"`
	RocketMQ RocketMQConfig `yaml:"rocketmq" json:"rocketmq" mapstructure:"rocketmq"`
	Pulsar   PulsarConfig   `yaml:"pulsar" json:"pulsar" mapstructure:"pulsar"`
}

// IOutput Defines the event output interface
type IOutput interface {
	// Send Sends an event to the downstream
	Send(ctx context.Context, event types.EventData) error
	// Close Closes the resource
	Close() error
}

func NewOutput(cfg Config) (IOutput, error) {
	// Look up the corresponding constructor function
	creator, exists := outputs[cfg.Type]
	if !exists {
		// Default to Stdout output
		return NewStdoutOutput()
	}
	// Call the constructor function to create the output instance
	return creator(cfg)
}

// SendWithRetry Sends with retry functionality
func SendWithRetry(ctx context.Context, output IOutput, event types.EventData, maxRetries int) error {
	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		if err := output.Send(ctx, event); err == nil {
			return nil
		} else {
			lastErr = err
			if i < maxRetries {
				time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
			}
		}
	}
	return lastErr
}
