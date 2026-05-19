package log

import (
	"log/slog"
	"sync"

	"go.uber.org/zap"
)

var (
	// globalAtomicLevel is the atomic level enabler for the global logger.
	globalAtomicLevel zap.AtomicLevel
	initOnce          sync.Once
)

// Init initializes the global logger with the provided options.
func Init(opts *Options) error {
	var err error
	initOnce.Do(func() {
		if opts == nil {
			opts = DefaultOptions()
		}

		if opts.FilePath != "" {
			opts.Output = OutputFile
		}

		// Initialize the atomic level with the configured level
		globalAtomicLevel = zap.NewAtomicLevelAt(toZapLevel(opts.Level))

		// Create the handler with the atomic level
		handler := NewZapHandler(opts, globalAtomicLevel)

		// Set the default slog logger
		slog.SetDefault(slog.New(handler))
	})
	return err
}

// New returns a fresh slog.Logger backed by the zap handler configured by
// opts. Unlike Init, New does not touch the process-global slog default or
// the globalAtomicLevel — it is safe to call multiple times and from tests
// that need isolated loggers.
//
// The returned logger ignores any sync.Once held by Init and has its own
// per-instance level (the level is captured at construction time; use Init +
// SetLevel for dynamic global level adjustment, or wrap a zap.AtomicLevel
// yourself).
func New(opts *Options) *slog.Logger {
	if opts == nil {
		opts = DefaultOptions()
	}
	if opts.FilePath != "" {
		opts.Output = OutputFile
	}
	level := zap.NewAtomicLevelAt(toZapLevel(opts.Level))
	return slog.New(NewZapHandler(opts, level))
}

// SetLevel dynamically sets the log level.
func SetLevel(l slog.Level) {
	globalAtomicLevel.SetLevel(toZapLevel(l))
}
