package output

import (
	"chihqiang/dbxgo/pkg/structx"
	"chihqiang/dbxgo/types"
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"time"
)

// KafkaConfig Kafka configuration entity, used to initialize KafkaOutput
type KafkaConfig struct {
	// Brokers List of Kafka brokers, e.g., ["127.0.0.1:9092"]
	Brokers []string `yaml:"brokers" json:"brokers" mapstructure:"brokers" env:"OUTPUT_KAFKA_BROKERS" envDefault:"127.0.0.1:9092"`

	// Topic The name of the Kafka topic to send messages to
	Topic string `yaml:"topic" json:"topic" mapstructure:"topic" env:"OUTPUT_KAFKA_TOPIC" envDefault:"dbxgo-events"`
}

// KafkaOutput Kafka implementation that satisfies the IOutput interface
type KafkaOutput struct {
	// writer Kafka writer
	writer *kafka.Writer
	// config Kafka configuration entity
	config KafkaConfig
}

// NewKafkaOutput Creates a KafkaOutput using the configuration entity
// Parameters:
//
//	cfg: KafkaConfig configuration struct, contains broker list and topic name
//
// Returns:
//
//	*KafkaOutput instance
func NewKafkaOutput(cfg KafkaConfig) (*KafkaOutput, error) {
	var (
		err error
	)
	cfg, err = structx.MergeWithDefaults[KafkaConfig](cfg)
	if err != nil {
		return nil, err
	}
	// Create Kafka writer
	writer := &kafka.Writer{
		// Kafka broker address list
		Addr: kafka.TCP(cfg.Brokers...),
		// Kafka topic to send messages to
		Topic: cfg.Topic,
		// Partition selection strategy, LeastBytes means selecting the partition with the least load
		Balancer: &kafka.LeastBytes{},
		// Wait for all replicas to confirm the message has been written, ensuring message reliability
		RequiredAcks: kafka.RequireAll,
		// Whether the send is asynchronous, false means synchronous sending
		Async: false,
	}
	return &KafkaOutput{
		writer: writer,
		config: cfg,
	}, nil
}

// Send Serializes EventData to a JSON string and sends it to Kafka
// Parameters:
//
//	event: types.EventData the event to send
//
// Returns:
//
//	error error if sending fails, otherwise nil
func (k *KafkaOutput) Send(ctx context.Context, event types.EventData) error {
	// Serialize the event into a JSON string
	eventValue, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return k.writer.WriteMessages(ctx, kafka.Message{
		Value: eventValue,
		Time:  time.Now(),
	})
}

// Close Closes the Kafka connection
// Returns:
//
//	error error if closing fails, otherwise nil
func (k *KafkaOutput) Close() error {
	return k.writer.Close()
}
