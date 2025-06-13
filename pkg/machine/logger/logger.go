// internal/logger/logger.go
package logger

import (
	"io"
	"log"
	"os"
)

type Logger struct {
	debug bool
	l     *log.Logger
}

func New(debug bool, logFile string) *Logger {
	f, err := os.Create(logFile)
	if err != nil {
		log.Printf("cannot create log file %s: %v — fallback to stdout only", logFile, err)
		f = os.Stdout
	}

	var w io.Writer = f
	if debug {
		w = io.MultiWriter(f, os.Stdout)
	}

	base := log.New(w, "", 0)

	return &Logger{
		debug: debug,
		l:     base,
	}
}

func NewForTest(w io.Writer) *Logger {
	l := New(true, "")
	l.l.SetOutput(w)
	return l
}

func (lg *Logger) Debug(v ...any) {
	lg.l.Print(
		append([]any{""}, v...)...,
	)
}
func (lg *Logger) Debugf(format string, v ...any) {
	lg.l.Printf(""+format, v...)
}

func (lg *Logger) Info(v ...any) {
	w := lg.l.Writer()
	lg.l.SetOutput(os.Stdout)
	lg.l.Print(append([]any{"INFO"}, v...)...)
	lg.l.SetOutput(w)
}

func (lg *Logger) Infof(format string, v ...any) {
	w := lg.l.Writer()
	lg.l.SetOutput(os.Stdout)
	lg.l.Printf(""+format, v...)
	lg.l.SetOutput(w)
}

func (lg *Logger) Error(v ...any) {
	lg.l.Print(
		append([]any{"ERROR "}, v...)...,
	)
}
func (lg *Logger) Errorf(format string, v ...any) {
	lg.l.Printf("ERROR "+format, v...)
}
