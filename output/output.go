package output

import (
	"context"
	"github.com/chihqiang/dbxgo/types"
	"time"
)

type OutputType string

var (
	OutputTypeStdout   OutputType = "stdout"
	OutputTypeRedis    OutputType = "redis"
	OutputTypeKafka    OutputType = "kafka"
	OutputTypeRabbitMQ OutputType = "rabbitmq"
	OutputTypeRocketMQ OutputType = "rocketmq"
	outputs                       = map[OutputType]func(Config) (IOutput, error){
		OutputTypeStdout:   func(cfg Config) (IOutput, error) { return NewStdoutOutput() },
		OutputTypeRedis:    func(cfg Config) (IOutput, error) { return NewRedisOutput(cfg.Redis) },
		OutputTypeKafka:    func(cfg Config) (IOutput, error) { return NewKafkaOutput(cfg.Kafka) },
		OutputTypeRabbitMQ: func(cfg Config) (IOutput, error) { return NewRabbitMQOutput(cfg.RabbitMQ) },
		OutputTypeRocketMQ: func(cfg Config) (IOutput, error) { return NewRocketMQOutput(cfg.RocketMQ) },
	}
)

func Register(outputType OutputType, fn func(Config) (IOutput, error)) {
	outputs[outputType] = fn
}

type Config struct {
	Type     OutputType     `yaml:"type"`
	Redis    RedisConfig    `yaml:"redis"`
	Kafka    KafkaConfig    `yaml:"kafka"`
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq"`
	RocketMQ RocketMQConfig `yaml:"rocketmq"`
}

// IOutput 定义事件输出接口
type IOutput interface {
	// Send 发送事件到下游
	Send(ctx context.Context, event types.EventData) error
	// Close 关闭资源
	Close() error
}

func NewOutput(cfg Config) (IOutput, error) {
	// 查找对应的构造函数
	creator, exists := outputs[cfg.Type]
	if !exists {
		// 默认为Stdout输出
		return NewStdoutOutput()
	}
	// 调用构造函数创建输出实例
	return creator(cfg)
}

// SendWithRetry 带重试的发送函数
func SendWithRetry(ctx context.Context, output IOutput, event types.EventData, maxRetries int) error {
	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		if err := output.Send(ctx, event); err == nil {
			return nil
		} else {
			lastErr = err
			if i < maxRetries {
				time.Sleep(time.Duration(i+1) * 100 * time.Millisecond) // 指数退避
			}
		}
	}
	return lastErr
}
