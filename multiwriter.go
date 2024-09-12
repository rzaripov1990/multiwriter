package multiwriter

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
)

const (
	ColorRed    = 31
	ColorYellow = 33
	ColorBlue   = 36
	ColorGray   = 37
)

type (
	MultiWriter struct {
		colorize  bool
		ignoreErr bool
		writers   []io.Writer
	}
)

func New(colorize bool, ignoreErrors bool, wr ...io.Writer) *MultiWriter {
	return &MultiWriter{
		colorize:  colorize,
		ignoreErr: ignoreErrors,
		writers:   wr,
	}
}

func (m *MultiWriter) Write(p []byte) (n int, err error) {
	color := func() int {
		if m.colorize {
			return levelColor[m.detectLogLevel(p)]
		}
		return 0
	}()

	for i := range m.writers {
		if m.writers[i] != nil {
			if m.colorize {
				if val, ok := m.writers[i].(interface{ Colored() bool }); ok && val.Colored() {
					n, err = fmt.Fprintf(m.writers[i], "\x1b[%dm%s\x1b[0m", color, string(p))
					if err != nil && !m.ignoreErr {
						return
					}
					continue
				}
			}

			n, err = m.writers[i].Write(p)
			if err != nil && !m.ignoreErr {
				return
			}
		}
	}
	return
}

var (
	stringLevel = map[string]slog.Level{
		`"INFO"`:  slog.LevelInfo,
		`=INFO`:   slog.LevelInfo,
		`"WARN"`:  slog.LevelWarn,
		`=WARN`:   slog.LevelWarn,
		`"ERROR"`: slog.LevelError,
		`=ERROR`:  slog.LevelError,
		`"DEBUG"`: slog.LevelDebug,
		`=DEBUG`:  slog.LevelDebug,
	}
	levelColor = map[slog.Level]int{
		slog.LevelInfo:  ColorBlue,
		slog.LevelDebug: ColorGray,
		slog.LevelWarn:  ColorYellow,
		slog.LevelError: ColorRed,
	}
)

func (m *MultiWriter) detectLogLevel(s []byte) slog.Level {
	for key, level := range stringLevel {
		if bytes.Index(s, []byte(key)) > 0 {
			return level
		}
	}
	return slog.LevelDebug
}
