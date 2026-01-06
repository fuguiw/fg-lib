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

// SetLevel dynamically sets the log level.
func SetLevel(l slog.Level) {
	globalAtomicLevel.SetLevel(toZapLevel(l))
}
