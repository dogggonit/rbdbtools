package logger

import (
	"fmt"
	"log"
	"os"
)

type Logger struct {
	warning *log.Logger
	info    *log.Logger
	error   *log.Logger
	fatal   *log.Logger
}

func New() Logger {
	opts := log.Ldate | log.Ltime
	return Logger{
		warning: log.New(os.Stdout, "WARNING: ", opts),
		info:    log.New(os.Stdout, "INFO: ", opts),
		error:   log.New(os.Stderr, "ERROR: ", opts),
		fatal:   log.New(os.Stderr, "FATAL: ", opts),
	}
}

func (l *Logger) Info(v ...interface{}) {
	l.info.Println(v)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.Info(fmt.Sprintf(format, v...))
}

func (l *Logger) Warning(v ...interface{}) {
	l.warning.Println(v)
}

func (l *Logger) Warningf(format string, v ...interface{}) {
	l.Warning(fmt.Sprintf(format, v...))
}

func (l *Logger) Error(v ...interface{}) {
	l.error.Println(v)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Error(fmt.Sprintf(format, v...))
}

func (l *Logger) Fatal(v ...interface{}) {
	l.fatal.Fatalln(v)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Fatal(fmt.Sprintf(format, v...))
}
