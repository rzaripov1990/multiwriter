package multiwriter

import (
	"context"
	"log"
	"log/slog"
	colored_logger "multiwriter/ext/colored"
	kafka_logger "multiwriter/ext/kafka"
	"os"
	"testing"

	kgo "github.com/segmentio/kafka-go"
)

func TestStd(t *testing.T) {
	logger := slog.New(
		slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{
				AddSource: false,
				Level:     slog.LevelDebug,
			},
		),
	)
	logger.Debug("start", "json", "true")
}

func TestMulti(t *testing.T) {
	{
		// writer #1
		file, err := os.Create("./custom-file.log") //create a new file
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

		colorize := true
		ignoreErrors := true

		logger := slog.New(
			slog.NewJSONHandler(
				New(
					colorize,
					ignoreErrors,
					stdout, // #1
					file,   // #2
					kafka,  // #3
					//os.Stdout, // #4 standart
				),
				&slog.HandlerOptions{
					AddSource: false,
					Level:     slog.LevelDebug,
				},
			),
		)

		var (
			ctx        context.Context = context.Background()
			slogValues []slog.Attr     = []slog.Attr{
				slog.String("possession", "version"),
				slog.String("song", "concept"),
				slog.String("construction", "direction"),
				slog.String("reading", "quantity"),
				slog.String("historian", "efficiency"),
				slog.String("establishment", "courage"),
			}
		)

		logger.LogAttrs(ctx, slog.LevelDebug, "Lorem Ipsum is simply dummy text of the printing and typesetting industry.", slogValues[:2]...)
		logger.LogAttrs(ctx, slog.LevelInfo, "Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book", slogValues[2:4]...)
		logger.LogAttrs(ctx, slog.LevelWarn, "It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged", slogValues[4:]...)
		logger.LogAttrs(ctx, slog.LevelError, "It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum.")
	}
}

func TestLog(t *testing.T) {
	file, err := os.Create("./custom-file.log")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	mw := New(false, false, os.Stdout, file)

	log.SetOutput(mw)
	log.Println(`"INFO": информационное сообщение`)
	log.Println(`"WARN": предупреждающее сообщение`)
	log.Println(`"ERROR": ошибка`)
	log.Println(`"DEBUG": сообщение для отладки`)
}
