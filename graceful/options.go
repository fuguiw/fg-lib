package graceful

import "time"

type Option func(*Graceful)

func WithTimeout(d time.Duration) Option {
	return func(g *Graceful) {
		g.timeout = d
	}
}
