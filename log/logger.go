package log

import (
	"runtime"
	"strings"
)

type Level int8

const (
	NoLevel Level = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	DPanicLevel
	PanicLevel
	FatalLevel
)

type Logger interface {
	Debug(msg ...string)
	Info(msg ...string)
	Warn(msg ...string)
	Error(msg ...string)
	Dpanic(msg ...string)
	Panic(msg ...string)
	Fatal(msg ...string)
}

func level(s string) Level {
	s = strings.ToLower(s)
	switch s {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "dpanic":
		return DPanicLevel
	case "panic":
		return PanicLevel
	case "fatal":
		return FatalLevel
	default:
		return NoLevel
	}
}

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case DPanicLevel:
		return "DPANIC"
	case PanicLevel:
		return "PANIC"
	case FatalLevel:
		return "FATAL"
	default:
		return "NOLEVEL"
	}
}

func caller(skip int) (pc uintptr, file string, line int, ok bool) {
	pc, file, line, ok = runtime.Caller(skip)
	return
}
