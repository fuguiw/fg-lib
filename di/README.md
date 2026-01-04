# DI Library

`di` is a simple dependency injection container wrapper based on `uber-go/dig`.

## Features

- **Simple API**: `Provide` and `Invoke`.
- **Singleton Management**: Ensures dependencies are instantiated once.
- **Named Dependencies**: Support for named instances.
- **Panic on Error**: Helper methods `MustProvide` and `MustInvoke` for easier initialization.

## Usage

```go
package main

import (
	"fmt"
	"github.com/fuguiw/fg-lib/di"
)

type Config struct {
	AppName string
}

func NewConfig() *Config {
	return &Config{AppName: "MyService"}
}

type Service struct {
	Conf *Config
}

func NewService(c *Config) *Service {
	return &Service{Conf: c}
}

func main() {
	container := di.GetDigContainer()

	// Register constructors
	container.MustProvide(NewConfig)
	container.MustProvide(NewService)

	// Invoke function to use dependencies
	container.MustInvoke(func(s *Service) {
		fmt.Printf("Service running for app: %s\n", s.Conf.AppName)
	})
}
```
