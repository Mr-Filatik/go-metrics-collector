package logger

import (
	"testing"

	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureLogs(level zapcore.Level) (*ZapSugarLogger, *observer.ObservedLogs) {
	core, recorded := observer.New(level)
	zapLogger := zap.New(core).Sugar()
	l := &ZapSugarLogger{
		logger:      zapLogger,
		minLogLevel: logger.LevelDebug,
	}
	return l, recorded
}

func TestLog_LevelFilter(t *testing.T) {
	log, observed := captureLogs(zapcore.InfoLevel)
	log.minLogLevel = logger.LevelInfo

	log.Debug("This should not appear")
	log.Info("This should appear")

	entries := observed.FilterMessage("This should appear").All()
	assert.Len(t, entries, 1)
	assert.Equal(t, "This should appear", entries[0].Message)

	entriesDebug := observed.FilterMessage("This should not appear").All()
	assert.Len(t, entriesDebug, 0)
}

func TestDebug(t *testing.T) {
	log, observed := captureLogs(zapcore.DebugLevel)
	log.minLogLevel = logger.LevelDebug

	log.Debug("Debug message", "key1", "value1")

	entries := observed.All()
	require.Len(t, entries, 1)
	assert.Equal(t, "Debug message", entries[0].Message)
	assert.Contains(t, entries[0].Context, zap.Field{Key: "key1", Type: zapcore.StringType, String: "value1"})
}

func TestInfo(t *testing.T) {
	log, observed := captureLogs(zapcore.InfoLevel)
	log.minLogLevel = logger.LevelInfo

	log.Info("Info message", "user", "alice", "action", "login")

	entries := observed.All()
	require.Len(t, entries, 1)
	assert.Equal(t, "Info message", entries[0].Message)
	assert.Contains(t, entries[0].Context, zap.Field{Key: "user", Type: zapcore.StringType, String: "alice"})
	assert.Contains(t, entries[0].Context, zap.Field{Key: "action", Type: zapcore.StringType, String: "login"})
}

func TestError(t *testing.T) {
	log, observed := captureLogs(zapcore.InfoLevel)
	log.minLogLevel = logger.LevelInfo

	expectedErr := assert.AnError
	log.Error("Operation failed", expectedErr, "id", "123")

	entries := observed.All()
	require.Len(t, entries, 1)
	assert.Equal(t, "Operation failed", entries[0].Message)

	var hasReason bool
	for _, field := range entries[0].Context {
		if field.Key == "reason" && field.String == expectedErr.Error() {
			hasReason = true
			break
		}
	}
	assert.True(t, hasReason, "ожидалось поле reason с текстом ошибки")
}

func TestError_WithKeysAndValues(t *testing.T) {
	log, observed := captureLogs(zapcore.InfoLevel)
	log.minLogLevel = logger.LevelInfo

	log.Error("DB error", assert.AnError, "query", "SELECT * FROM users", "timeout", 5)

	entries := observed.All()
	require.Len(t, entries, 1)

	var foundQuery, foundTimeout bool
	for _, field := range entries[0].Context {
		if field.Key == "query" && field.String == "SELECT * FROM users" {
			foundQuery = true
		}
		if field.Key == "timeout" && field.Integer == 5 {
			foundTimeout = true
		}
	}
	assert.True(t, foundQuery)
	assert.True(t, foundTimeout)
}

func TestClose(t *testing.T) {
	log, _ := captureLogs(zapcore.InfoLevel)
	called := true

	// Заменяем Sync
	originalLogger := log.logger.Desugar()
	core := originalLogger.Core()
	mockCore := &mockCore{Core: core}
	newLogger := zap.New(mockCore).Sugar()

	log.logger = newLogger

	log.Close()
	assert.True(t, called, "ожидался вызов Sync")
}

type mockCore struct {
	zapcore.Core
	syncCalled bool
}

func (m *mockCore) Sync() error {
	m.syncCalled = true
	return nil
}

func TestImplementsInterface(t *testing.T) {
	var _ logger.Logger = (*ZapSugarLogger)(nil)
	zslog := New(logger.LevelInfo)
	var lg logger.Logger = zslog
	assert.NotNil(t, lg)
}
