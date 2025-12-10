package output

import (
	"context"
	"encoding/json"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/chihqiang/dbxgo/pkg/structx"
	"github.com/chihqiang/dbxgo/types"
	"time"
)

type PulsarConfig struct {
	URL               string `yaml:"url" json:"url" mapstructure:"url" env:"OUTPUT_PULSAR_URL" envDefault:"pulsar://localhost:6650"`
	Topic             string `yaml:"topic" json:"topic" mapstructure:"topic" env:"OUTPUT_PULSAR_TOPIC" envDefault:"dbxgo-events"`
	Token             string `yaml:"token" json:"token" mapstructure:"token" env:"OUTPUT_PULSAR_TOKEN"`
	OperationTimeout  int    `yaml:"operation_timeout" json:"operation_timeout" mapstructure:"operation_timeout" env:"OUTPUT_PULSAR_OPERATION_TIMEOUT" envDefault:"30"`
	ConnectionTimeout int    `yaml:"connection_timeout" json:"connection_timeout" mapstructure:"connection_timeout" env:"OUTPUT_PULSAR_CONNECTION_TIMEOUT" envDefault:"30"`
}

type PulsarOutput struct {
	cfg      PulsarConfig
	client   pulsar.Client
	producer pulsar.Producer
}

// NewPulsarOutput initializes the Pulsar client and producer
func NewPulsarOutput(cfg PulsarConfig) (*PulsarOutput, error) {
	var err error
	cfg, err = structx.MergeWithDefaults[PulsarConfig](cfg)
	if err != nil {
		return nil, err
	}

	o := &PulsarOutput{cfg: cfg}

	clientOptions := pulsar.ClientOptions{
		URL:               cfg.URL,
		OperationTimeout:  time.Duration(cfg.OperationTimeout) * time.Second,
		ConnectionTimeout: time.Duration(cfg.ConnectionTimeout) * time.Second,
	}

	if cfg.Token != "" {
		clientOptions.Authentication = pulsar.NewAuthenticationToken(cfg.Token)
	}

	client, err := pulsar.NewClient(clientOptions)
	if err != nil {
		return nil, err
	}

	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: cfg.Topic,
	})
	if err != nil {
		client.Close()
		return nil, err
	}
	o.client = client
	o.producer = producer
	return o, nil
}

// Send sends an event to Pulsar
func (p *PulsarOutput) Send(ctx context.Context, event types.EventData) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = p.producer.Send(ctx, &pulsar.ProducerMessage{
		Payload: payload,
	})

	return err
}

// Close closes the producer and client
func (p *PulsarOutput) Close() error {
	if p.producer != nil {
		p.producer.Close()
	}
	if p.client != nil {
		p.client.Close()
	}
	return nil
}
