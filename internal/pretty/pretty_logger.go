package pretty

import (
	"io"
	"runtime"
	"time"
)

type PrettyLogger interface {
	PrettyErrorLog(w io.Writer)
}

type LoggerConfig struct {
	Prefix string `json:"prefix"`
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

func (p *prettyPrinter) PrettyErrorLog() {
	logType := "ERROR"
	timeStamp := time.Now()
	timeStampString := p.DateTimeSting(timeStamp)
	curFunc, line := callerFunction()
}
