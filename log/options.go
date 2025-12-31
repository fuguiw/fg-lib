package log

import (
	"log/slog"
)

type Format string

const (
	FormatJSON Format = "json"
	FormatText Format = "text"
)

type Output string

const (
	OutputStdout Output = "stdout"
	OutputStderr Output = "stderr"
	OutputFile   Output = "file"
)

type Options struct {
	// Level is the minimum log level to output.
	Level slog.Level

	// Format specifies the output format (json or text).
	Format Format

	// Output specifies where to write logs (stdout, stderr, or file).
	Output Output

	// FilePath is the path to the log file (required if Output is "file").
	FilePath string

	// Rotation configuration for file output.
	MaxSize    int  // megabytes
	MaxBackups int  // number of files
	MaxAge     int  // days
	Compress   bool // compress rotated files

	// EnableCaller enables caller reporting.
	EnableCaller bool

	// EnableStack enables stack trace recording.
	EnableStack bool
}

func DefaultOptions() *Options {
	return &Options{
		Level:        slog.LevelInfo,
		Format:       FormatJSON,
		Output:       OutputStdout,
		EnableCaller: true,
		EnableStack:  true,
		MaxSize:      10,
		MaxBackups:   4,
		MaxAge:       28,
		Compress:     true,
	}
}
