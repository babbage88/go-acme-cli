package pretty

import (
	"io"
	"runtime"
)

type CustomLogger struct {
	level  string
	output io.Writer
}

type CustomLoggerPrefixProperty struct {
	Value     string `json:"value"`
	Index     int8   `json:"index"`
	Padding   int8   `json:"padding"`
	Seperator string `json:"seperator"`
}

func NewCustomLogger(output io.Writer, level string) *CustomLogger {
	return &CustomLogger{
		level:  level,
		output: output,
	}
}

func (l *CustomLogger) Info(message string) {
	if l.level == "info" || l.level == "warn" || l.level == "error" {
		l.output.Write([]byte("[INFO] " + message + "\n"))
	}
}

func getCurrentFileLine() (string, int) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return "unknown", 0
	}
	return file, line
}

func callerFunction() (string, int) {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return "unknown", 0
	}
	funcName := runtime.FuncForPC(pc).Name()
	return funcName, 0
}

func currentFunction() (string, int) {
	pc, _, _, ok := runtime.Caller(0)
	if !ok {
		return "unknown", 0
	}
	funcName := runtime.FuncForPC(pc).Name()
	return funcName, 0
}
