# fg-lib

`fg-lib` is a lightweight Go utility library providing robust Logging and Configuration management.

## Features

### Log (`github.com/fuguiw/fg-lib/log`)
A structured logging wrapper based on `log/slog` and `zap`.

- **Standard API**: Built on top of `log/slog`.
- **High Performance**: Powered by `uber-go/zap`.
- **Flexible Configuration**:
  - Formats: JSON, Text.
  - Outputs: Stdout, Stderr, File (with rotation).
  - Levels: Dynamic level adjustment.
- **Context Aware**: Automatic injection/extraction of `trace_id`, `request_id`, `user_id`.
- **Rich Details**: Configurable caller and stacktrace reporting.

### Config (`github.com/fuguiw/fg-lib/config`)
A struct-based configuration loader.

- **Source Priority**: Env Vars > File > Defaults.
- **Tag Support**: `default`, `yaml`, `json`, `env`.
- **Auto Refresh**: Watch for changes and reload automatically.
- **Zero Dependency**: Minimal external dependencies (only `yaml.v3` for parsing).

## Installation

```bash
go get github.com/fuguiw/fg-lib
```

## Usage

### Logger

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

### Configuration

```go
package main

import (
	"fmt"
	"time"

	"github.com/fuguiw/fg-lib/config"
)

type AppConfig struct {
	Name string `yaml:"name" default:"my-app" env:"APP_NAME"`
	Port int    `yaml:"port" default:"8080" env:"APP_PORT"`
}

func main() {
	var cfg AppConfig
	
	// Load from file (optional) and env
	loader := config.NewLoader(&cfg, config.WithFile("config.yaml"))
	if err := loader.Load(); err != nil {
		panic(err)
	}

	fmt.Printf("Config: %+v\n", cfg)

	// Watch for updates
	loader.StartAutoRefresh(10*time.Second, func(newCfg interface{}) {
		updated := newCfg.(*AppConfig)
		fmt.Printf("Updated: %+v\n", updated)
	})
	
	select {}
}
```

