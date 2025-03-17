package logger

type LogLevel uint32

const (
	LevelDebug LogLevel = 1
	LevelInfo  LogLevel = 2
)

type Logger interface {
	Log(level LogLevel, message string, keysAndValues ...interface{})
	Debug(message string, keysAndValues ...interface{})
	Info(message string, keysAndValues ...interface{})
	Close()
}
