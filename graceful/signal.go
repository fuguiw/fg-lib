package graceful

import (
	"os"
	"os/signal"
	"syscall"
)

func setupSignal(g *Graceful) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-ch
		g.cancel()
	}()
}
