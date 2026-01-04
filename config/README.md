# Config Library

`config` is a lightweight, struct-based configuration loader for Go.

## Features

- **Struct-Based Loading**: Define your configuration using Go structs.
- **Multiple Sources**: Loads from Defaults, Files (YAML/JSON), and Environment Variables.
- **Priority**: Environment Variables > File > Defaults.
- **Auto Refresh**: Watch for changes and reload automatically.
- **Tag Support**:
  - `default`: Set default values.
  - `yaml` / `json`: Map file keys.
  - `env`: Map environment variables.

## Usage

```go
package main

import (
	"fmt"
	"time"

	"github.com/fuguiw/fg-lib/config"
)

type DatabaseConfig struct {
	Host string `yaml:"host" default:"localhost" env:"DB_HOST"`
	Port int    `yaml:"port" default:"3306" env:"DB_PORT"`
}

type AppConfig struct {
	Name     string         `yaml:"name" default:"my-app" env:"APP_NAME"`
	Debug    bool           `yaml:"debug" default:"false" env:"APP_DEBUG"`
	Database DatabaseConfig `yaml:"database"`
}

func main() {
	var cfg AppConfig
	
	// Initialize Loader with config file
	loader := config.NewLoader(&cfg, config.WithFile("config.yaml"))

	// Load Config
	if err := loader.Load(); err != nil {
		panic(err)
	}

	fmt.Printf("Initial Config: %+v\n", cfg)

	// Watch for updates (e.g. file changes)
	loader.StartAutoRefresh(10*time.Second, func(newCfg interface{}) {
		updated := newCfg.(*AppConfig)
		fmt.Printf("Config Updated: %+v\n", updated)
	})
	
	select {}
}
```
