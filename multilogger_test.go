package multiwriter

import (
	"log"
	"log/slog"
	colored_logger "multiwriter/ext/colored"
	kafka_logger "multiwriter/ext/kafka"
	"os"
	"testing"

	kgo "github.com/segmentio/kafka-go"
)

func TestStd(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: false}))
	logger.Debug("start", "json", "true")
}

func TestMulti(t *testing.T) {
	// writer #1
	file, err := os.Create("./myfile.txt") //create a new file
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// writer #2
	kafka := kafka_logger.New(
		&kgo.Writer{
			Addr:                   kgo.TCP("127.0.0.1:9092"),
			Balancer:               &kgo.RoundRobin{},
			RequiredAcks:           kgo.RequireOne,
			Async:                  true,
			Compression:            kgo.Compression(0),
			AllowAutoTopicCreation: true,
			Topic:                  "logs",
		},
	)
	defer kafka.Close()

	// writer #3
	stdout := colored_logger.New(true, os.Stdout) // wrap os.Stdout for colorize

	colorize := false
	ignoreErrors := true

	logger := slog.New(
		slog.NewJSONHandler(
			New(
				colorize,
				ignoreErrors,
				stdout,    // #1
				file,      // #2
				kafka,     // #3
				os.Stdout, // #4
			),
			&slog.HandlerOptions{
				AddSource: false,
				Level:     slog.LevelDebug,
			},
		),
	)

	logger.Debug("start", "json", "true")
	logger.Info("start", "json", "true")
	logger.Warn("start", "json", "true")
	logger.Error("start", "json", "true")

	log.Print("sssss")
}

// type (
// 	JsonKafkaStdOut struct {
// 		mw MultiWriter
// 	}
// )

// func NewJsonKafkaStdOut(wr ...io.Writer) *JsonKafkaStdOut {
// 	return &JsonKafkaStdOut{
// 		mw: *NewMultiWriter(wr...),
// 	}
// }

// func (h *JsonKafkaStdOut) Enabled(_ context.Context, level slog.Level) bool {
// 	return false
// }
// func (h *JsonKafkaStdOut) Handle(_ context.Context, r slog.Record) error {
// 	return nil
// }
// func (h *JsonKafkaStdOut) WithAttrs(attrs []slog.Attr) slog.Handler {
// 	return h
// }
// func (h *JsonKafkaStdOut) WithGroup(name string) slog.Handler {
// 	return h
// }
