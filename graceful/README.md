# Graceful Library

`graceful` provides a mechanism for graceful shutdown of applications.

## Features

- **Signal Handling**: Listens for `SIGINT` and `SIGTERM`.
- **Component Management**: Manages lifecycle of multiple components.
- **Timeout Control**: Configurable shutdown timeout (default 15s).
- **Context Propagation**: Passes cancellation context to components.

## Usage

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/fuguiw/fg-lib/graceful"
)

// Server simulates a long-running service
type Server struct{}

func (s *Server) Run(ctx context.Context) error {
	fmt.Println("Server started")
	<-ctx.Done() // Wait for shutdown signal
	fmt.Println("Server stopping...")
	time.Sleep(1 * time.Second) // Simulate cleanup
	fmt.Println("Server stopped")
	return nil
}

func main() {
	// Initialize graceful manager
	g := graceful.New(graceful.WithTimeout(5 * time.Second))

	// Register components
	server := &Server{}
	g.Register(server)

	// Start components and wait for signal
	if err := g.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
```
