# Log Library

`log` is a structured logging wrapper based on `log/slog` and `zap`.

## Features

- **Standard API**: Built on top of `log/slog`.
- **High Performance**: Powered by `uber-go/zap`.
- **Flexible Configuration**:
  - Formats: JSON, Text.
  - Outputs: Stdout, Stderr, File (with rotation).
  - Levels: Dynamic level adjustment.
- **Context Aware**: Automatic injection/extraction of `trace_id`, `request_id`, `user_id`.
- **Rich Details**: Configurable caller and stacktrace reporting.

## Usage

```go
package main

import (
	"context"
	"log/slog"

	"github.com/fuguiw/fg-lib/log"
)

func main() {
	// Initialize
	opts := log.DefaultOptions()
	opts.Level = slog.LevelDebug
	opts.Format = log.FormatJSON
	
	if err := log.Init(opts); err != nil {
		panic(err)
	}

	// Use slog as usual
	slog.Info("Hello World", "key", "value")

	// Context with Trace ID
	ctx := log.WithTraceID(context.Background(), "trace-123")
	slog.InfoContext(ctx, "Processing request") // output includes trace_id

	// Dynamic Level
	log.SetLevel(slog.LevelInfo)
}
```
