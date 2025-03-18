package ports

type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(error)
	Debug(msg string, args ...any)
}
