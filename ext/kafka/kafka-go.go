package kafka_logger

import (
	"context"

	kgo "github.com/segmentio/kafka-go"
)

type (
	KafkaLogger struct {
		ctx    context.Context
		client *kgo.Writer
	}
)

func (m *KafkaLogger) Write(p []byte) (n int, err error) {
	err = m.client.WriteMessages(m.ctx, kgo.Message{Value: p})
	return
}

func (m *KafkaLogger) Close() {
	m.client.Close()
}

func (m *KafkaLogger) Colored() bool {
	return false // always disable
}

func New(kafkaWriter *kgo.Writer) *KafkaLogger {
	return &KafkaLogger{
		ctx:    context.Background(),
		client: kafkaWriter,
	}
}
