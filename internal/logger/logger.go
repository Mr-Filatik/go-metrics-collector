package logger

type Logger interface {
	Info(message string, keysAndValues ...interface{})
	Close()
}
