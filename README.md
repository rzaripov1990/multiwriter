# MultiWriter

`multiwriter` is a Go library that allows logging to multiple destinations that implement the `io.Writer` interface, with optional colorization based on log levels. It is designed to handle simultaneous output to multiple sources, such as files, consoles, or external logging systems like Kafka.

## Features

- Log simultaneously to multiple destinations (`io.Writer`).
- Colorize output based on log levels (`INFO`, `WARN`, `ERROR`, `DEBUG`).
- Best-effort mode to ignore destination errors (`ignoreErrors=true`).
- Support for custom destinations, with opt-in color output capability detection (`Colored() bool`).

## Behavior

- `Write(p)` **returns `n=len(p)` on success** (even when colorization adds ANSI bytes to some destinations).
- If `ignoreErrors=false`, the first destination error stops the write and is returned.
- If `ignoreErrors=true`, destination errors are ignored and `Write(p)` returns **`(len(p), nil)`**.
- `MultiWriter` serializes concurrent `Write` calls (so log lines are less likely to interleave).

## Installation

Install the package using `go get`:

```bash
go get github.com/rzaripov1990/multiwriter
```

## Examples
### Basic Example: Log to Console and File

The following example demonstrates how to log to both the console (os.Stdout) and a log file:

```go
package main

import (
    "log"
    "os"
    "github.com/rzaripov1990/multiwriter"
)

func main() {
    file, err := os.Create("./custom-file.log")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    // Create a MultiWriter instance for console and file output
    mw := multiwriter.New(false, false, os.Stdout, file)

    // Set the logger output to MultiWriter
    log.SetOutput(mw)

    // Log messages with different levels
    log.Println(`"INFO": informational message`)
    log.Println(`"WARN": warning message`)
    log.Println(`"ERROR": error message`)
    log.Println(`"DEBUG": debug message`)
}
```

### Advanced Example: Log to Console and Kafka Topic

In this example, logs are sent to both the console (with colorization) and a Kafka topic:

```go
package main

import (
    "context"
    "os"

    colored_logger "github.com/rzaripov1990/multiwriter/ext/colored"
    kafka_logger "github.com/rzaripov1990/multiwriter/ext/kafka"
    "github.com/rzaripov1990/multiwriter"
    "log/slog"

    kgo "github.com/segmentio/kafka-go"
)

func main() {
    // Enable colorization and ignore errors
    colorize := true
    ignoreErrors := true

    // Destination 1: Kafka logger setup
    kafka := kafka_logger.New(
        &kgo.Writer{
            Addr:     kgo.TCP("127.0.0.1:9092"),
            Balancer: &kgo.RoundRobin{},
            Topic:    "logs",
        },
    )
    defer kafka.Close()

    // Destination 2: Standard output with colorization
    stdout := colored_logger.New(true, os.Stdout)

    // Create logger with MultiWriter
    logger := slog.New(
        slog.NewJSONHandler(
            multiwriter.New(colorize, ignoreErrors, stdout, kafka),
            &slog.HandlerOptions{
                AddSource: false,
                Level:     slog.LevelDebug,
            },
        ),
    )

    ctx := context.Background()
    slogValues := []slog.Attr{
        slog.String("possession", "version"),
        slog.String("song", "concept"),
        slog.String("construction", "direction"),
        slog.String("reading", "quantity"),
        slog.String("historian", "efficiency"),
        slog.String("establishment", "courage"),
    }

    // Log messages with various log levels and attributes
    logger.LogAttrs(ctx, slog.LevelDebug, "Debugging log message", slogValues[:2]...)
    logger.LogAttrs(ctx, slog.LevelInfo, "Informational log message", slogValues[2:4]...)
    logger.LogAttrs(ctx, slog.LevelWarn, "Warning log message", slogValues[4:]...)
    logger.LogAttrs(ctx, slog.LevelError, "Error log message", nil)
}
```

### Adding New Destinations

To add a new log destination, simply implement the `io.Writer` interface. Optionally, if the destination supports colorized output, implement the `Colored() bool` function.

```go
type CustomWriter struct {}

func (cw *CustomWriter) Write(p []byte) (n int, err error) {
    // Custom implementation for writing logs
    return len(p), nil
}

func (cw *CustomWriter) Colored() bool {
    return true // Supports colorized output
}
```

## Development

Run unit tests:

```bash
go test ./...
```

Run integration tests (Kafka/file demo) explicitly:

```bash
go test -tags=integration ./...
```

To bring up Kafka locally for integration tests, see `example/docker-compose.yml`.