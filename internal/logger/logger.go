// Пакет logger содержит абстрактное описание логгера используемого в проекте.
// Необходимо для лёгкой и быстрой замены одной реализации логгера на другую.
package logger

type LogLevel uint32

// Константы - уровни логирования.
const (
	LevelDebug LogLevel = 1 // уровень логирования debug
	LevelInfo  LogLevel = 2 // уровень логирования info
	LevelError LogLevel = 3 // уровень логирования error
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
