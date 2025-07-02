package domain

type LogProvider interface {
	Debug(msg string, data map[string]any)
	Info(msg string, data map[string]any)
	Warning(msg string, data map[string]any)
	Error(msg string, data map[string]any)
	Critical(msg string, data map[string]any)

	SetLogLevel(level LogLevel)
}
