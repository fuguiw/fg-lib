package graceful

import (
	"context"
	"errors"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockComponent struct {
	name      string
	startFunc func(context.Context) error
	stopFunc  func(context.Context) error
}

func (m *mockComponent) Name() string {
	return m.name
}

func (m *mockComponent) Start(ctx context.Context) error {
	if m.startFunc != nil {
		return m.startFunc(ctx)
	}
	return nil
}

func (m *mockComponent) Stop(ctx context.Context) error {
	if m.stopFunc != nil {
		return m.stopFunc(ctx)
	}
	return nil
}

func TestGraceful_Run(t *testing.T) {
	g := New()

	started := false
	stopped := false

	c := &mockComponent{
		name: "test",
		startFunc: func(ctx context.Context) error {
			started = true
			return nil
		},
		stopFunc: func(ctx context.Context) error {
			stopped = true
			return nil
		},
	}

	g.Register(c)

	go func() {
		time.Sleep(100 * time.Millisecond)
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	err := g.Run()
	assert.NoError(t, err)
	assert.True(t, started)
	assert.True(t, stopped)
}

func TestGraceful_StartError(t *testing.T) {
	g := New()
	expectedErr := errors.New("start failed")

	c := &mockComponent{
		name: "fail-start",
		startFunc: func(ctx context.Context) error {
			return expectedErr
		},
	}

	g.Register(c)

	err := g.Run()
	assert.Equal(t, expectedErr, err)
}

func TestGraceful_StopOrder(t *testing.T) {
	g := New()
	var order []string

	c1 := &mockComponent{
		name:      "c1",
		startFunc: func(ctx context.Context) error { return nil },
		stopFunc: func(ctx context.Context) error {
			order = append(order, "c1")
			return nil
		},
	}
	c2 := &mockComponent{
		name:      "c2",
		startFunc: func(ctx context.Context) error { return nil },
		stopFunc: func(ctx context.Context) error {
			order = append(order, "c2")
			return nil
		},
	}

	g.Register(c1, c2)

	go func() {
		time.Sleep(50 * time.Millisecond)
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	err := g.Run()
	assert.NoError(t, err)
	// Stop order should be reverse of registration: c2, c1
	assert.Equal(t, []string{"c2", "c1"}, order)
}

func TestGraceful_ShutdownTimeout(t *testing.T) {
	g := New(WithTimeout(100 * time.Millisecond))

	stopCalled := make(chan struct{})

	c := &mockComponent{
		name:      "slow-stop",
		startFunc: func(ctx context.Context) error { return nil },
		stopFunc: func(ctx context.Context) error {
			close(stopCalled)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(500 * time.Millisecond):
				return nil
			}
		},
	}

	g.Register(c)

	go func() {
		time.Sleep(50 * time.Millisecond)
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	start := time.Now()
	err := g.Run()
	duration := time.Since(start)

	assert.NoError(t, err)

	select {
	case <-stopCalled:
		// pass
	default:
		t.Fatal("Stop was not called")
	}

	// Should take at least 50ms (wait for signal) + 100ms (timeout)
	// But less than 50ms + 500ms (slow stop)
	assert.True(t, duration < 400*time.Millisecond, "shutdown took too long, timeout might not have worked")
}
