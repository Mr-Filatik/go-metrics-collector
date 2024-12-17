package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var GlobalLogger *zap.SugaredLogger

func Initialize(level zapcore.Level) {
	log, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	GlobalLogger = log.Sugar()
	if err := log.Sync(); err != nil {
		panic(err)
	}
}

func Debug(message string, keysAndValues ...interface{}) {
	checkLogger()
	GlobalLogger.Debugw(message, keysAndValues...)
}

func Info(message string, keysAndValues ...interface{}) {
	checkLogger()
	GlobalLogger.Infow(message, keysAndValues...)
}

func Warn(message string, keysAndValues ...interface{}) {
	checkLogger()
	GlobalLogger.Warnw(message, keysAndValues...)
}

func Error(err error, keysAndValues ...interface{}) {
	checkLogger()
	GlobalLogger.Errorw(err.Error(), keysAndValues...)
}

func checkLogger() {
	if GlobalLogger == nil {
		zap.L().Sugar().Warn("Logger not initialized")
	}
	Initialize(zapcore.InfoLevel)
}
