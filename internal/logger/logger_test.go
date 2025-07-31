package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLevelName(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected string
	}{
		{"Debug", LevelDebug, "debug"},
		{"Info", LevelInfo, "info"},
		{"Error", LevelError, "error"},
		{"Unknown low", 0, "none"},
		{"Unknown high", 999, "none"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetLevelName(tt.level)
			assert.Equal(t, tt.expected, result)
		})
	}
}

type compileMock struct{}

func (m *compileMock) Log(level LogLevel, message string, keysAndValues ...interface{}) {}
func (m *compileMock) Debug(message string, keysAndValues ...interface{})               {}
func (m *compileMock) Info(message string, keysAndValues ...interface{})                {}
func (m *compileMock) Error(message string, err error, keysAndValues ...interface{})    {}
func (m *compileMock) Close()                                                           {}

func TestLoggerInterface(t *testing.T) {
	var logger Logger = &compileMock{}
	if logger == nil {
		t.Fatal("mock не должен быть nil")
	}
}
