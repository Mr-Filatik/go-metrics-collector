// Пакет logger содержит абстрактное описание логгера используемого в проекте.
// Необходимо для лёгкой и быстрой замены одной реализации логгера на другую.
package logger

// LogLevel описывает уровень логирования.
type LogLevel uint32

// Константы - уровни логирования.
const (
	LevelDebug LogLevel = 1 // уровень логирования debug
	LevelInfo  LogLevel = 2 // уровень логирования info
	LevelError LogLevel = 3 // уровень логирования error
)

// Logger описывает интерфейс для всех логгеров используемых в проекте.
type Logger interface {
	Log(level LogLevel, message string, keysAndValues ...interface{}) // общий метод логирования
	Debug(message string, keysAndValues ...interface{})               // логирование с уровнем debug
	Info(message string, keysAndValues ...interface{})                // логирование с уровнем info
	Error(message string, err error, keysAndValues ...interface{})    // логирование с уровнем error
	Close()                                                           // закрытие ресурсов связанных с логгером
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
