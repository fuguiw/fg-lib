package graceful

import (
	"context"
	"sync"
	"time"
)

type Graceful struct {
	ctx    context.Context
	cancel context.CancelFunc

	timeout time.Duration

	comps []Component
	mu    sync.Mutex
}

func New(opts ...Option) *Graceful {
	ctx, cancel := context.WithCancel(context.Background())

	g := &Graceful{
		ctx:     ctx,
		cancel:  cancel,
		timeout: 15 * time.Second,
	}

	for _, opt := range opts {
		opt(g)
	}

	setupSignal(g)
	return g
}

func (g *Graceful) Context() context.Context {
	return g.ctx
}

func (g *Graceful) Register(cs ...Component) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.comps = append(g.comps, cs...)
}
