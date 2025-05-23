package logger

import (
	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
	"go.uber.org/zap"
)

type LogLevel = logger.LogLevel

const (
	LevelDebug = logger.LevelDebug
	LevelInfo  = logger.LevelInfo
)

type ZapSugarLogger struct {
	logger      *zap.SugaredLogger
	minLogLevel LogLevel
}

func New(minLogLevel LogLevel) *ZapSugarLogger {
	log, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zslog := &ZapSugarLogger{
		minLogLevel: minLogLevel,
		logger:      log.Sugar(),
	}
	zslog.Info(
		"Create logger",
		"name", "ZapSugarLogger",
		"level", logger.GetLevelName(zslog.minLogLevel),
	)
	return zslog
}

func (l *ZapSugarLogger) Log(level LogLevel, message string, keysAndValues ...interface{}) {
	if level >= l.minLogLevel {
		l.logger.Infow(message, keysAndValues...)
	}
}

func (l *ZapSugarLogger) Debug(message string, keysAndValues ...interface{}) {
	if LevelDebug >= l.minLogLevel {
		l.logger.Infow(message, keysAndValues...)
	}
}

func (l *ZapSugarLogger) Info(message string, keysAndValues ...interface{}) {
	if LevelInfo >= l.minLogLevel {
		l.logger.Infow(message, keysAndValues...)
	}
}

func (l *ZapSugarLogger) Error(message string, err error, keysAndValues ...interface{}) {
	if LevelInfo >= l.minLogLevel {
		addKeysAndValues := append([]interface{}{"reason", err.Error()}, keysAndValues...)
		l.logger.Infow(message, addKeysAndValues...)
	}
}

func (l *ZapSugarLogger) Close() {
	if err := l.logger.Sync(); err != nil {
		panic(err)
	}
}
