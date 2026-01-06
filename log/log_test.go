package log

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestLogToFile(t *testing.T) {
	// Create a temporary directory for logs
	tmpDir, err := os.MkdirTemp("", "log_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "test.log")

	// Initialize logger
	opts := DefaultOptions()
	opts.Output = OutputFile
	opts.FilePath = logFile
	opts.Format = FormatJSON
	opts.Level = slog.LevelInfo

	// We can't re-init the global logger easily because of sync.Once in Init.
	// So we will test NewZapHandler directly.
	handler := NewZapHandler(opts, zap.NewAtomicLevelAt(toZapLevel(opts.Level)))
	logger := slog.New(handler)

	// Write a log
	msg := "test log message to file"
	logger.Info(msg)

	// Sleep briefly to ensure write flush
	time.Sleep(100 * time.Millisecond)

	// Check file content
	content, err := os.ReadFile(logFile)
	assert.NoError(t, err)
	assert.True(t, len(content) > 0, "Log file should not be empty")
	assert.True(t, strings.Contains(string(content), msg), "Log file should contain the message")
}
