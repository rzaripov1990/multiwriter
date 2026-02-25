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
	// Copy p because kafka-go writer may retain the slice beyond the call
	// (especially with Async=true), while callers are allowed to reuse p.
	b := make([]byte, len(p))
	copy(b, p)

	err = m.client.WriteMessages(m.ctx, kgo.Message{Value: b})
	if err != nil {
		return 0, err
	}
	return len(p), nil
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
