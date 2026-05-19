package log

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/exp/zapslog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// TestReplaceAttr_Redacts verifies that Options.ReplaceAttr runs against
// every record attribute and that an empty-Key return drops the attr.
func TestReplaceAttr_Redacts(t *testing.T) {
	_ = lumberjack.Logger{} // keep the lumberjack import wired even when only direct zap APIs are used here

	var buf bytes.Buffer

	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.RFC3339TimeEncoder
	cfg.StacktraceKey = ""
	encoder := zapcore.NewJSONEncoder(cfg)
	level := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), level)
	inner := zapslog.NewHandler(core)

	opts := &Options{
		Level:  slog.LevelInfo,
		Format: FormatJSON,
		Output: OutputStdout,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == "api_key" {
				return slog.Attr{Key: a.Key, Value: slog.StringValue("***")}
			}
			if a.Key == "drop_me" {
				return slog.Attr{} // empty key → dropped
			}
			return a
		},
	}
	h := &ZapHandler{Handler: inner, opts: opts}
	logger := slog.New(h)

	logger.Info("test",
		slog.String("api_key", "should-not-appear"),
		slog.String("keep", "visible"),
		slog.String("drop_me", "should-not-appear-either"),
	)

	out := buf.String()
	var rec map[string]any
	if err := json.Unmarshal([]byte(out), &rec); err != nil {
		t.Fatalf("output is not JSON: %v\nraw: %s", err, out)
	}
	if rec["api_key"] != "***" {
		t.Errorf("api_key should be redacted to ***, got %v", rec["api_key"])
	}
	if rec["keep"] != "visible" {
		t.Errorf("keep field altered, got %v", rec["keep"])
	}
	if _, ok := rec["drop_me"]; ok {
		t.Errorf("drop_me should have been dropped, got %v", rec["drop_me"])
	}
	if strings.Contains(out, "should-not-appear") {
		t.Errorf("raw secret leaked into output: %s", out)
	}
}

// TestNew_StandaloneInstance verifies New returns an isolated logger that
// does not require Init's sync.Once.
func TestNew_StandaloneInstance(t *testing.T) {
	lg := New(&Options{Level: slog.LevelInfo, Format: FormatJSON, Output: OutputStdout})
	if lg == nil {
		t.Fatal("New returned nil")
	}
	lg.Info("works")
}
