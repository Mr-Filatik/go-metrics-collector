// Пакет logger содержит абстрактное описание логгера используемого в проекте.
// Необходимо для лёгкой и быстрой замены одной реализации логгера на другую.
package logger

// LogLevel описывает уровень логирования.
type LogLevel uint32

// Константы - уровни логирования.
const (
	LevelDebug LogLevel = iota // уровень логирования debug
	LevelInfo                  // уровень логирования info
	LevelWarn                  // уровень логирования warning
	LevelError                 // уровень логирования error
)

// Logger описывает интерфейс для всех логгеров используемых в проекте.
type Logger interface {
	// Log - универсальный метод логирования.
	Log(level LogLevel, message string, keysAndValues ...interface{})

	// Debug - логирование с уровнем debug.
	Debug(message string, keysAndValues ...interface{})

	// Info - логирование с уровнем info.
	Info(message string, keysAndValues ...interface{})

	// Warn - логирование с уровнем warn и возможной (некритичной) ошибкой.
	Warn(message string, err error, keysAndValues ...interface{})

	// Error - логирование с уровнем error и критичной ошибкой
	Error(message string, err error, keysAndValues ...interface{})

	// Close - закрытие ресурсов связанных с логгером
	Close()
}

// GetLevelName преобразует уровень логирования в строку.
//
// Параметры:
//   - logLevel: уровень логирования.
func GetLevelName(logLevel LogLevel) string {
	switch logLevel {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warning"
	case LevelError:
		return "error"
	default:
		return "none"
	}
}
