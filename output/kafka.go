package output

import (
	"context"
	"encoding/json"
	"github.com/chihqiang/dbxgo/pkg/structsx"
	"github.com/chihqiang/dbxgo/types"
	"github.com/segmentio/kafka-go"
	"time"
)

// KafkaConfig Kafka 配置实体，用于初始化 KafkaOutput
type KafkaConfig struct {
	// Brokers Kafka broker 列表，例如 ["127.0.0.1:9092"]
	Brokers []string `yaml:"brokers" json:"brokers" mapstructure:"brokers" env:"OUTPUT_KAFKA_BROKERS" envDefault:"127.0.0.1:9092"`

	// Topic 要发送的 Kafka topic 名称
	Topic string `yaml:"topic" json:"topic" mapstructure:"topic" env:"OUTPUT_KAFKA_TOPIC" envDefault:"dbxgo-events"`
}

// KafkaOutput Kafka 实现，满足 IOutput 接口
type KafkaOutput struct {
	// writer Kafka 写入器
	writer *kafka.Writer
	// config Kafka 配置实体
	config KafkaConfig
}

// NewKafkaOutput 使用配置实体创建 KafkaOutput
// 参数:
//
//	cfg: KafkaConfig 配置结构体，包含 broker 列表和 topic 名称
//
// 返回:
//
//	*KafkaOutput 实例
func NewKafkaOutput(cfg KafkaConfig) (*KafkaOutput, error) {
	var (
		err error
	)
	cfg, err = structsx.MergeWithDefaults[KafkaConfig](cfg)
	if err != nil {
		return nil, err
	}
	// 创建 Kafka writer
	writer := &kafka.Writer{
		// Kafka broker 地址列表
		Addr: kafka.TCP(cfg.Brokers...),
		// 消息发送到的 Kafka topic
		Topic: cfg.Topic,
		// 分区选择策略，LeastBytes 表示选择当前负载最小的分区
		Balancer: &kafka.LeastBytes{},
		// 等待所有副本确认消息已写入，保证消息可靠性
		RequiredAcks: kafka.RequireAll,
		// 是否异步发送，false 表示同步发送
		Async: false,
	}
	return &KafkaOutput{
		writer: writer,
		config: cfg,
	}, nil
}

// Send 将 EventData 序列化为 JSON 字符串并发送到 Kafka
// 参数:
//
//	event: types.EventData 要发送的事件
//
// 返回:
//
//	error 发送失败时返回错误，否则为 nil
func (k *KafkaOutput) Send(ctx context.Context, event types.EventData) error {
	// 将事件序列化为 JSON 字符串
	eventValue, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return k.writer.WriteMessages(ctx, kafka.Message{
		Value: eventValue,
		Time:  time.Now(),
	})
}

// Close 关闭 Kafka 连接
// 返回:
//
//	error 关闭失败时返回错误，否则为 nil
func (k *KafkaOutput) Close() error {
	return k.writer.Close()
}
