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

	// "" и 0 → без времени/даты, только то, что передаём в Print/Printf.
	base := log.New(w, "", 0)

	return &Logger{
		debug: debug,
		l:     base,
	}
}

func (lg *Logger) Debug(v ...any) {
	lg.l.Print(
		//TODO: вернуть
		append([]any{"DEBUG "}, v...)...,
	// append([]any{""}, v...)...,
	)
}

func (lg *Logger) Debugf(format string, v ...any) {
	lg.l.Printf("DEBUG "+format, v...)
}

func (lg *Logger) Info(v ...any) {
	lg.l.Print(append([]any{"INFO"}, v...)...)
}

func (lg *Logger) Infof(format string, v ...any) {
	lg.l.Printf("INFO "+format, v...)
}
