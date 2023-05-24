package lkit_go

import (
	"github.com/rs/zerolog"
	"github.com/xlkness/lkit-go/internal/log"
)

type Handler = log.Handler
type LogLevel = log.LogLevel

func NewGlobalLogger(writer Handler, level LogLevel, initFun func(logger zerolog.Logger) zerolog.Logger, terminalDebug bool) {
	log.NewGlobalLogger(writer, level, initFun, terminalDebug)
}

func Tracef(format string, v ...interface{}) {
	log.Tracef(format, v...)
}

func Debugf(format string, v ...interface{}) {
	log.Debugf(format, v...)
}

func Infof(format string, v ...interface{}) {
	log.Infof(format, v...)
}

func Noticef(format string, v ...interface{}) {
	log.Noticef(format, v...)
}

func Warnf(format string, v ...interface{}) {
	log.Warnf(format, v...)
}

func Errorf(format string, v ...interface{}) {
	log.Errorf(format, v...)
}

func Critif(format string, v ...interface{}) {
	log.Critif(format, v...)
}

func Fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}
