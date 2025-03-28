package logger

type LogLevel uint8

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarning
	LevelError
)

type Logger interface {
	Debug(message string, keysAndValues ...interface{})
	Info(message string, keysAndValues ...interface{})
	Warning(message string, keysAndValues ...interface{})
	Error(message string, err error, keysAndValues ...interface{})
	Close()
}

func GetLevelName(logLevel LogLevel) string {
	switch logLevel {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarning:
		return "warning"
	case LevelError:
		return "error"
	default:
		return "none"
	}
}
