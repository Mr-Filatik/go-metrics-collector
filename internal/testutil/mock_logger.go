package testutil

import (
	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
)

type MockLoggerLog struct {
	Keyvals []interface{}
	Err     error
	Message string
	Level   logger.LogLevel
}

type MockLogger struct {
	logs []MockLoggerLog
}

func (m *MockLogger) Log(log logger.LogLevel, msg string, keyvals ...interface{}) {
	m.logs = append(m.logs, MockLoggerLog{
		Level:   log,
		Message: msg,
		Err:     nil,
		Keyvals: keyvals,
	})
}

func (m *MockLogger) Debug(msg string, keyvals ...interface{}) {
	m.logs = append(m.logs, MockLoggerLog{
		Level:   logger.LevelDebug,
		Message: msg,
		Err:     nil,
		Keyvals: keyvals,
	})
}

func (m *MockLogger) Info(msg string, keyvals ...interface{}) {
	m.logs = append(m.logs, MockLoggerLog{
		Level:   logger.LevelInfo,
		Message: msg,
		Err:     nil,
		Keyvals: keyvals,
	})
}

func (m *MockLogger) Warn(msg string, err error, keyvals ...interface{}) {
	m.logs = append(m.logs, MockLoggerLog{
		Level:   logger.LevelWarn,
		Message: msg,
		Err:     err,
		Keyvals: keyvals,
	})
}

func (m *MockLogger) Error(msg string, err error, keyvals ...interface{}) {
	m.logs = append(m.logs, MockLoggerLog{
		Level:   logger.LevelError,
		Message: msg,
		Err:     err,
		Keyvals: keyvals,
	})
}

func (m *MockLogger) Close() {}

// GetAllLogs выдаёт весь набор логов из MockLogger.
func (m *MockLogger) GetAllLogs() []MockLoggerLog {
	return m.logs
}

// GetAllLogsByLevel выдаёт весь набор логов из MockLogger для к определённого уровня логирования.
func (m *MockLogger) GetAllLogsByLevel(level logger.LogLevel) []MockLoggerLog {
	var filtered []MockLoggerLog
	for i := range m.logs {
		if m.logs[i].Level == level {
			filtered = append(filtered, m.logs[i])
		}
	}
	return filtered
}

// GetLastLog выдаёт последний лог из MockLogger.
func (m *MockLogger) GetLastLog() *MockLoggerLog {
	if len(m.logs) == 0 {
		return nil
	}
	return &m.logs[len(m.logs)-1]
}

// GetLastLogByLevel выдаёт последний лог из MockLogger для определённого уровня логирования.
func (m *MockLogger) GetLastLogByLevel(level logger.LogLevel) *MockLoggerLog {
	var filtered []MockLoggerLog
	for i := range m.logs {
		if m.logs[i].Level == level {
			filtered = append(filtered, m.logs[i])
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return &filtered[len(filtered)-1]
}
