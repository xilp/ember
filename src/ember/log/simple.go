package log

import (
	"bufio"
	"io"
	"fmt"
	"time"
)

func (p *Log) Debug(msg string) {
	p.output(LogLevelDebug, msg)
}

func (p *Log) Info(msg string) {
	p.output(LogLevelInfo, msg)
}

func (p *Log) Warn(msg string) {
	p.output(LogLevelWarn, msg)
}

func (p *Log) Error(msg string) {
	p.output(LogLevelError, msg)
}

func (p *Log) output(level int, msg string) {
	if level < p.level {
		return
	}
	str := fmt.Sprintf("[%s] %s %s", time.Now().Format("2006-01-02 15:04:05"), LogLevel(level).String(), msg)
	p.w.Write([]byte(str))
}

func (p *Log) Close() {
	p.w.Flush()
}

func (p *Log) SetLevel(level int) {
	p.level = level
}

func NewLog(w io.Writer, level int) *Log {
	return &Log{bufio.NewWriter(w), level}
}

type Log struct {
	w *bufio.Writer
	level int
}

func (p LogLevel) String() string {
	switch p {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	}
	return "UNKNOWN"
}

type LogLevel int

const (
	LogLevelDebug = 0
	LogLevelInfo  = 1
	LogLevelWarn  = 2
	LogLevelError = 3
)
