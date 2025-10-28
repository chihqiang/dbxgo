package output

import (
	"context"
	"encoding/json"
	"fmt"
	"chihqiang/dbxgo/pkg/structsx"
	"chihqiang/dbxgo/types"
	"github.com/rabbitmq/amqp091-go"
	"time"
)

// RabbitMQConfig RabbitMQ configuration entity
type RabbitMQConfig struct {
	URL        string `yaml:"url" json:"url" mapstructure:"url" env:"OUTPUT_RABBITMQ_URL" envDefault:"amqp://guest:guest@127.0.0.1:5672/"`
	Exchange   string `yaml:"exchange" json:"exchange" mapstructure:"exchange" env:"OUTPUT_RABBITMQ_EXCHANGE" envDefault:"dbxgo-exchange"`
	Queue      string `yaml:"queue" json:"queue" mapstructure:"queue" env:"OUTPUT_RABBITMQ_QUEUE" envDefault:"dbxgo-events"`
	Durable    bool   `yaml:"durable" json:"durable" mapstructure:"durable" env:"OUTPUT_RABBITMQ_DURABLE" envDefault:"true"`
	AutoDelete bool   `yaml:"auto_delete" json:"auto_delete" mapstructure:"auto_delete" env:"OUTPUT_RABBITMQ_AUTODELETE" envDefault:"false"`
	AutoAck    bool   `yaml:"auto_ack" json:"auto_ack" mapstructure:"auto_ack" env:"OUTPUT_RABBITMQ_AUTOACK" envDefault:"false"`
	Exclusive  bool   `yaml:"exclusive" json:"exclusive" mapstructure:"exclusive" env:"OUTPUT_RABBITMQ_EXCLUSIVE" envDefault:"false"`
	NoWait     bool   `yaml:"no_wait" json:"no_wait" mapstructure:"no_wait" env:"OUTPUT_RABBITMQ_NOWAIT" envDefault:"false"`
}

// RabbitMQOutput RabbitMQ output implementation
type RabbitMQOutput struct {
	config RabbitMQConfig
	conn   *amqp091.Connection
	ch     *amqp091.Channel
}

// NewRabbitMQOutput Creates a RabbitMQOutput and tests the connection
func NewRabbitMQOutput(cfg RabbitMQConfig) (*RabbitMQOutput, error) {
	var (
		err error
	)
	cfg, err = structsx.MergeWithDefaults[RabbitMQConfig](cfg)
	if err != nil {
		return nil, err
	}
	// Establish RabbitMQ connection
	conn, err := amqp091.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	// Open a channel
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}
	// Declare the queue
	_, err = ch.QueueDeclare(
		cfg.Queue,
		cfg.Durable,
		cfg.AutoDelete,
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

// Send Serializes EventJSON to a JSON string and sends it to RabbitMQ
func (r *RabbitMQOutput) Send(ctx context.Context, event types.EventData) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	return r.ch.PublishWithContext(ctx,
		r.config.Exchange,
		r.config.Queue,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		},
	)
}

// Close Closes the RabbitMQ connection
func (r *RabbitMQOutput) Close() error {
	if r.ch != nil {
		_ = r.ch.Close()
	}
	if r.conn != nil {
		_ = r.conn.Close()
	}
	return nil
}
