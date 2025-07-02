package domain

import "log"

type LogLevel int

const (
	Debug LogLevel = iota
	Info
	Warning
	Error
	Critical
)

func ParseLogLevel(level string) LogLevel {
	switch level {
	case "DEBUG":
		return Debug
	case "INFO":
		return Info
	case "WARNING":
		return Warning
	case "ERROR":
		return Error
	case "CRITICAL":
		return Critical
	default:
		log.Printf("Log level '%s' not recognized, using INFO by default", level)
		return Info // Nivel por defecto
	}
}
