package logger

type LogLevel uint32

const (
	LevelDebug LogLevel = 1
	LevelInfo  LogLevel = 2
	LevelError LogLevel = 3
)

type Logger interface {
	Log(level LogLevel, message string, keysAndValues ...interface{})
	Debug(message string, keysAndValues ...interface{})
	Info(message string, keysAndValues ...interface{})
	Error(message string, err error, keysAndValues ...interface{})
	Close()
}

func GetLevelName(logLevel LogLevel) string {
	switch logLevel {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelError:
		return "error"
	default:
		return "none"
	}
}
