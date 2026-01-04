package graceful

import (
	"context"
)

func (g *Graceful) Run() error {
	for _, c := range g.comps {
		if err := c.Start(g.ctx); err != nil {
			return err
		}
	}

	<-g.ctx.Done()

	return g.stopAll()
}

func (g *Graceful) stopAll() error {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()

	for i := len(g.comps) - 1; i >= 0; i-- {
		_ = g.comps[i].Stop(ctx)
	}
	return nil
}
