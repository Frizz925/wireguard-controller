package logger

import (
	"fmt"
	"io"
)

const INDENT_SIZE = 2

type Logger struct {
	level  int
	output io.Writer
}

func New(w io.Writer) *Logger {
	return &Logger{
		output: w,
	}
}

func (l *Logger) Indent() *Logger {
	return &Logger{
		level:  l.level + 1,
		output: l.output,
	}
}

func (l *Logger) Log(format string, a ...any) {
	l.writeIndent()
	fmt.Fprintf(l.output, format, a...)
	fmt.Fprintln(l.output)
}

func (l *Logger) writeIndent() {
	for i := 0; i < l.level*INDENT_SIZE; i++ {
		fmt.Fprintf(l.output, " ")
	}
}
