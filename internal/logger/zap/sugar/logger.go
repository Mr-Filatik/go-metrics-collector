package logger

import "go.uber.org/zap"

type ZapSugarLogger struct {
	logger *zap.SugaredLogger
}

func New() *ZapSugarLogger {
	log, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zslog := ZapSugarLogger{
		logger: log.Sugar(),
	}
	zslog.Info(
		"Create logger",
		"name", "ZapSugarLogger",
		"level", "info",
	)
	return &zslog
}

func (l *ZapSugarLogger) Info(message string, keysAndValues ...interface{}) {
	l.logger.Infow(message, keysAndValues...)
}

func (l *ZapSugarLogger) Close() {
	if err := l.logger.Sync(); err != nil {
		panic(err)
	}
}
