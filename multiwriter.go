package multiwriter

import (
	"bytes"
	"io"
	"log/slog"
	"strconv"
	"sync"
)

const (
	ColorRed    = 31
	ColorYellow = 33
	ColorBlue   = 36
	ColorGray   = 37
)

type (
	MultiWriter struct {
		mu        sync.Mutex
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
	m.mu.Lock()
	defer m.mu.Unlock()

	color := 0
	if m.colorize {
		color = levelColor[m.detectLogLevel(p)]
	}

	for _, w := range m.writers {
		if w == nil {
			continue
		}

		if m.colorize {
			if val, ok := w.(interface{ Colored() bool }); ok && val.Colored() {
				if werr := writeColored(w, color, p); werr != nil && !m.ignoreErr {
					return 0, werr
				}
				continue
			}
		}

		wn, werr := w.Write(p)
		if werr == nil && wn != len(p) {
			werr = io.ErrShortWrite
		}
		if werr != nil && !m.ignoreErr {
			return wn, werr
		}
	}

	// If configured to ignore destination errors, report success to the caller
	// (so loggers don't treat the write as failed).
	if m.ignoreErr {
		return len(p), nil
	}
	return len(p), nil
}

var (
	levelColor = map[slog.Level]int{
		slog.LevelInfo:  ColorBlue,
		slog.LevelDebug: ColorGray,
		slog.LevelWarn:  ColorYellow,
		slog.LevelError: ColorRed,
	}
	levelPatterns = []struct {
		needle []byte
		level  slog.Level
	}{
		{[]byte(`"ERROR"`), slog.LevelError},
		{[]byte(`=ERROR`), slog.LevelError},
		{[]byte(`"WARN"`), slog.LevelWarn},
		{[]byte(`=WARN`), slog.LevelWarn},
		{[]byte(`"INFO"`), slog.LevelInfo},
		{[]byte(`=INFO`), slog.LevelInfo},
		{[]byte(`"DEBUG"`), slog.LevelDebug},
		{[]byte(`=DEBUG`), slog.LevelDebug},
	}
)

func (m *MultiWriter) detectLogLevel(s []byte) slog.Level {
	for i := range levelPatterns {
		if bytes.Contains(s, levelPatterns[i].needle) {
			return levelPatterns[i].level
		}
	}
	return slog.LevelDebug
}

func writeColored(w io.Writer, color int, p []byte) error {
	// \x1b[<color>m + payload + \x1b[0m
	buf := make([]byte, 0, len(p)+16)
	buf = append(buf, 0x1b, '[')
	buf = strconv.AppendInt(buf, int64(color), 10)
	buf = append(buf, 'm')
	buf = append(buf, p...)
	buf = append(buf, 0x1b, '[', '0', 'm')

	n, err := w.Write(buf)
	if err == nil && n != len(buf) {
		return io.ErrShortWrite
	}
	return err
}
