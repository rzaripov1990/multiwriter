package multiwriter

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

type coloredBuffer struct {
	bytes.Buffer
	colored bool
}

func (c *coloredBuffer) Colored() bool { return c.colored }

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, errors.New("boom") }

type coloredErrWriter struct{}

func (coloredErrWriter) Colored() bool             { return true }
func (coloredErrWriter) Write([]byte) (int, error) { return 0, errors.New("boom") }

type shortWriter struct {
	n int
}

func (s shortWriter) Write(p []byte) (int, error) {
	if s.n > len(p) {
		return len(p), nil
	}
	return s.n, nil
}

func TestMultiWriter_WritesToAllAndReturnsLen(t *testing.T) {
	var a, b bytes.Buffer
	mw := New(false, false, &a, &b)

	p := []byte("hello\n")
	n, err := mw.Write(p)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if n != len(p) {
		t.Fatalf("expected n=len(p)=%d, got %d", len(p), n)
	}
	if got := a.String(); got != string(p) {
		t.Fatalf("buffer A mismatch: %q", got)
	}
	if got := b.String(); got != string(p) {
		t.Fatalf("buffer B mismatch: %q", got)
	}
}

func TestMultiWriter_IgnoreErrors_ReturnsSuccess(t *testing.T) {
	var ok bytes.Buffer
	mw := New(false, true, errWriter{}, &ok)

	p := []byte("hello\n")
	n, err := mw.Write(p)
	if err != nil {
		t.Fatalf("expected nil err, got: %v", err)
	}
	if n != len(p) {
		t.Fatalf("expected n=len(p)=%d, got %d", len(p), n)
	}
	if got := ok.String(); got != string(p) {
		t.Fatalf("expected ok writer to receive payload, got: %q", got)
	}
}

func TestMultiWriter_NoIgnoreErrors_StopsOnError(t *testing.T) {
	var ok bytes.Buffer
	mw := New(false, false, errWriter{}, &ok)

	p := []byte("hello\n")
	n, err := mw.Write(p)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if n != 0 {
		t.Fatalf("expected n=0, got %d", n)
	}
	if got := ok.Len(); got != 0 {
		t.Fatalf("expected subsequent writers not called, got %d bytes", got)
	}
}

func TestMultiWriter_Colorize_WritesANSIAndReturnsLen(t *testing.T) {
	cw := &coloredBuffer{colored: true}
	mw := New(true, false, cw)

	// This payload matches the README-style log prefix, and also covers the
	// "match at position 0" case for level detection.
	p := []byte(`"INFO" test` + "\n")

	n, err := mw.Write(p)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if n != len(p) {
		t.Fatalf("expected n=len(p)=%d, got %d", len(p), n)
	}

	out := cw.Bytes()
	if !bytes.HasPrefix(out, []byte("\x1b[36m")) { // ColorBlue
		t.Fatalf("expected blue ANSI prefix, got: %q", string(out))
	}
	if !bytes.HasSuffix(out, []byte("\x1b[0m")) {
		t.Fatalf("expected ANSI reset suffix, got: %q", string(out))
	}
	if !bytes.Contains(out, p) {
		t.Fatalf("expected output to contain original payload, got: %q", string(out))
	}
}

func TestMultiWriter_Colorize_IgnoreErrorsStillSucceeds(t *testing.T) {
	var ok bytes.Buffer
	mw := New(true, true, coloredErrWriter{}, &ok)

	p := []byte(`"WARN" test` + "\n")
	n, err := mw.Write(p)
	if err != nil {
		t.Fatalf("expected nil err, got: %v", err)
	}
	if n != len(p) {
		t.Fatalf("expected n=len(p)=%d, got %d", len(p), n)
	}
	if got := ok.String(); got != string(p) {
		t.Fatalf("expected ok writer to receive raw payload, got: %q", got)
	}
}

func TestMultiWriter_ShortWriteBecomesErrShortWrite(t *testing.T) {
	var ok bytes.Buffer
	mw := New(false, false, shortWriter{n: 2}, &ok)

	p := []byte("hello\n")
	n, err := mw.Write(p)
	if !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("expected io.ErrShortWrite, got: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected n=2, got %d", n)
	}
	if ok.Len() != 0 {
		t.Fatalf("expected subsequent writers not called on short write")
	}
}
