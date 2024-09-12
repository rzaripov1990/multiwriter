package colored_logger

import (
	"io"
)

type (
	ColoredLogger struct {
		colorize bool
		file     io.Writer
	}
)

func (m *ColoredLogger) Write(p []byte) (n int, err error) {
	n, err = m.file.Write(p)
	return
}

func (m *ColoredLogger) Close() {}

func (m *ColoredLogger) Colored() bool {
	return m.colorize
}

func New(colorize bool, writer io.Writer) *ColoredLogger {
	return &ColoredLogger{
		colorize: colorize,
		file:     writer,
	}
}
