package log

import (
	"context"
	"io"
	"log/slog"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type ZapHandler struct {
	slog.Handler
	opts *Options
}

func NewZapHandler(opts *Options, level zap.AtomicLevel) *ZapHandler {
	var encoder zapcore.Encoder

	if opts.Format == FormatJSON {
		cfg := zap.NewProductionEncoderConfig()
		cfg.EncodeTime = zapcore.RFC3339TimeEncoder
		if !opts.EnableStack {
			cfg.StacktraceKey = ""
		}
		encoder = zapcore.NewJSONEncoder(cfg)
	} else {
		cfg := zap.NewDevelopmentEncoderConfig()
		cfg.EncodeTime = zapcore.RFC3339TimeEncoder
		if !opts.EnableStack {
			cfg.StacktraceKey = ""
		}
		encoder = zapcore.NewConsoleEncoder(cfg)
	}

	var writer io.Writer
	switch opts.Output {
	case OutputStdout:
		writer = os.Stdout
	case OutputStderr:
		writer = os.Stderr
	case OutputFile:
		writer = &lumberjack.Logger{
			Filename:   opts.FilePath,
			MaxSize:    opts.MaxSize,
			MaxBackups: opts.MaxBackups,
			MaxAge:     opts.MaxAge,
			Compress:   opts.Compress,
			LocalTime:  true,
		}
	default:
		writer = os.Stdout
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(writer),
		level,
	)

	// zapslog options
	slHandler := zapslog.NewHandler(core, zapslog.WithCaller(opts.EnableCaller))

	return &ZapHandler{
		Handler: slHandler,
		opts:    opts,
	}
}

// Handle overrides the underlying handler's Handle method to inject context
// fields and (optionally) apply Options.ReplaceAttr.
//
// zapslog does not expose slog.HandlerOptions.ReplaceAttr to its callers, so
// we apply the hook here before delegating to the zap backend. When the hook
// is nil we keep the original fast path (add ctx fields, delegate).
func (h *ZapHandler) Handle(ctx context.Context, record slog.Record) error {
	fields := contextFields(ctx)

	if h.opts.ReplaceAttr == nil {
		if len(fields) > 0 {
			record.AddAttrs(fields...)
		}
		return h.Handler.Handle(ctx, record)
	}

	// ReplaceAttr path: rebuild the record so each attr passes through the
	// hook. An attr returned with an empty Key is dropped (matches slog's
	// standard ReplaceAttr contract).
	rebuilt := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	record.Attrs(func(a slog.Attr) bool {
		if out := h.opts.ReplaceAttr(nil, a); out.Key != "" {
			rebuilt.AddAttrs(out)
		}
		return true
	})
	for _, a := range fields {
		if out := h.opts.ReplaceAttr(nil, a); out.Key != "" {
			rebuilt.AddAttrs(out)
		}
	}
	return h.Handler.Handle(ctx, rebuilt)
}

func toZapLevel(l slog.Level) zapcore.Level {
	switch {
	case l < slog.LevelInfo:
		return zapcore.DebugLevel
	case l < slog.LevelWarn:
		return zapcore.InfoLevel
	case l < slog.LevelError:
		return zapcore.WarnLevel
	default:
		return zapcore.ErrorLevel
	}
}
