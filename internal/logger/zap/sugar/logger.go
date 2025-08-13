// Пакет logger содержит конкретную реализацию логгера.
// Данный логгер использует простую версию zap логгера.
package logger

import (
	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
	"go.uber.org/zap"
)

// LogLevel описывает уровень логирования.
type LogLevel = logger.LogLevel

// Константы - уровни логирования.
const (
	LevelDebug = logger.LevelDebug // уровень логирования debug
	LevelInfo  = logger.LevelInfo  // уровень логирования info
	LevelWarn  = logger.LevelWarn  // уровень логирования warning
	LevelError = logger.LevelError // уровень логирования error
)

// ZapSugarLogger хранит информацию о логгере.
type ZapSugarLogger struct {
	logger      *zap.SugaredLogger // ссылка на реализацию логгера
	minLogLevel LogLevel           // минимальный уровень логирования
}

var _ logger.Logger = (*ZapSugarLogger)(nil)

// New инициализирует и создаёт новый экземпляр *ZapSugarLogger.
//
// Параметры:
//   - minLogLevel: минимальный уровень логирования
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

// Log логирует сообщение и параметры.
//
// Параметры:
//   - level: уровень лога
//   - message: сообщение
//   - keysAndValues: дополнительные пары ключ-значение
func (l *ZapSugarLogger) Log(level LogLevel, message string, keysAndValues ...interface{}) {
	if level >= l.minLogLevel {
		l.logger.Infow(message, keysAndValues...)
	}
}

// Debug логирует сообщение и параметры в уровне debug.
//
// Параметры:
//   - message: сообщение
//   - keysAndValues: дополнительные пары ключ-значение
func (l *ZapSugarLogger) Debug(message string, keysAndValues ...interface{}) {
	if LevelDebug >= l.minLogLevel {
		l.logger.Infow(message, keysAndValues...)
	}
}

// Info логирует сообщение и параметры в уровне info.
//
// Параметры:
//   - message: сообщение
//   - keysAndValues: дополнительные пары ключ-значение
func (l *ZapSugarLogger) Info(message string, keysAndValues ...interface{}) {
	if LevelInfo >= l.minLogLevel {
		l.logger.Infow(message, keysAndValues...)
	}
}

// Warn логирует сообщение и параметры в уровне warning.
// Дополнительно логирует причину ошибки в поле reason.
//
// Параметры:
//   - message: сообщение
//   - err: ошибка
//   - keysAndValues: дополнительные пары ключ-значение
func (l *ZapSugarLogger) Warn(message string, err error, keysAndValues ...interface{}) {
	if LevelInfo >= l.minLogLevel {
		addKeysAndValues := append([]interface{}{"reason", err.Error()}, keysAndValues...)
		l.logger.Infow(message, addKeysAndValues...)
	}
}

// Error логирует сообщение и параметры в уровне error.
// Дополнительно логирует причину ошибки в поле reason.
//
// Параметры:
//   - message: сообщение
//   - err: ошибка
//   - keysAndValues: дополнительные пары ключ-значение
func (l *ZapSugarLogger) Error(message string, err error, keysAndValues ...interface{}) {
	if LevelInfo >= l.minLogLevel {
		addKeysAndValues := append([]interface{}{"reason", err.Error()}, keysAndValues...)
		l.logger.Infow(message, addKeysAndValues...)
	}
}

// Close освобождает все используемые логгером ресурсы.
func (l *ZapSugarLogger) Close() {
	if err := l.logger.Sync(); err != nil {
		// На Windows есть ошибка при вызове Sync, поэтому игнорирую ошибку.
		// Все ресурсы должны освободиться в любом случае.
		_ = err
	}
}
