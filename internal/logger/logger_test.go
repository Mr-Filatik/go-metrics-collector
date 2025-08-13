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
		{"Info", LevelWarn, "warning"},
		{"Error", LevelError, "error"},
		{"Unknown", 999, "none"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetLevelName(tt.level)
			assert.Equal(t, tt.expected, result)
		})
	}
}
