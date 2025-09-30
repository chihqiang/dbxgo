package output

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chihqiang/dbxgo/types"
	"github.com/rabbitmq/amqp091-go"
	"time"
)

// RabbitMQConfig RabbitMQ 配置实体
type RabbitMQConfig struct {
	URL       string `yaml:"url"`       // 连接 URL，例如 amqp://guest:guest@127.0.0.1:5672/
	Queue     string `yaml:"queue"`     // 队列名
	Durable   bool   `yaml:"durable"`   // 队列是否持久化
	AutoAck   bool   `yaml:"auto_ack"`  // 是否自动确认
	Exclusive bool   `yaml:"exclusive"` // 是否排他队列
	NoWait    bool   `yaml:"no_wait"`   // 是否等待声明完成
}

// DefaultRabbitMQConfig 返回默认配置
func DefaultRabbitMQConfig() RabbitMQConfig {
	return RabbitMQConfig{
		URL:       "amqp://guest:guest@127.0.0.1:5672/",
		Queue:     "dbxgo",
		Durable:   true,
		AutoAck:   false,
		Exclusive: false,
		NoWait:    false,
	}
}

// RabbitMQOutput RabbitMQ 输出实现
type RabbitMQOutput struct {
	config RabbitMQConfig
	conn   *amqp091.Connection
	ch     *amqp091.Channel
}

// NewRabbitMQOutput 创建 RabbitMQOutput，并测试连接
func NewRabbitMQOutput(cfg RabbitMQConfig) (*RabbitMQOutput, error) {
	// 获取默认配置
	def := DefaultRabbitMQConfig()

	// 对每个字段进行判断，如果未设置则使用默认值
	if cfg.URL == "" {
		cfg.URL = def.URL
	}
	if cfg.Queue == "" {
		cfg.Queue = def.Queue
	}
	// 布尔值需要区分零值，通常 false 是合理默认
	// 这里 Durable、AutoAck、Exclusive、NoWait 可以直接使用 cfg 的值
	// 如果想强制默认值覆盖零值，可以加条件：if !cfg.Durable { cfg.Durable = def.Durable } 等

	// 建立 RabbitMQ 连接
	conn, err := amqp091.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// 打开 Channel
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// 声明队列
	_, err = ch.QueueDeclare(
		cfg.Queue,
		cfg.Durable,
		false, // autoDelete
		cfg.Exclusive,
		cfg.NoWait,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &RabbitMQOutput{
		config: cfg,
		conn:   conn,
		ch:     ch,
	}, nil
}

// Send 将 EventJSON 序列化为 JSON 字符串并发送到 RabbitMQ
func (r *RabbitMQOutput) Send(ctx context.Context, event types.EventData) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	return r.ch.Publish(
		"",             // exchange
		r.config.Queue, // routing key
		false,          // mandatory
		false,          // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		},
	)
}

// Close 关闭 RabbitMQ 连接
func (r *RabbitMQOutput) Close() error {
	if r.ch != nil {
		_ = r.ch.Close()
	}
	if r.conn != nil {
		_ = r.conn.Close()
	}
	return nil
}
