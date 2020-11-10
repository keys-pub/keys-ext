package server

import (
	pkglog "log"
)

// Logger compatible with GCP.
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// LogLevel ...
type LogLevel int

const (
	// DebugLevel ...
	DebugLevel LogLevel = 3
	// InfoLevel ...
	InfoLevel LogLevel = 2
	// WarnLevel ...
	WarnLevel LogLevel = 1
	// ErrLevel ...
	ErrLevel LogLevel = 0
	// NoLevel ...
	NoLevel LogLevel = -1
)

// NewLogger ...
func NewLogger(lev LogLevel) Logger {
	return &defaultLog{Level: lev}
}

type defaultLog struct {
	Level LogLevel
}

func (l defaultLog) Debugf(format string, args ...interface{}) {
	if l.Level >= 3 {
		pkglog.Printf("[DEBG] "+format+"\n", args...)
	}
}

func (l defaultLog) Infof(format string, args ...interface{}) {
	if l.Level >= 2 {
		pkglog.Printf("[INFO] "+format+"\n", args...)
	}
}

func (l defaultLog) Warningf(format string, args ...interface{}) {
	if l.Level >= 1 {
		pkglog.Printf("[WARN] "+format+"\n", args...)
	}
}

func (l defaultLog) Errorf(format string, args ...interface{}) {
	if l.Level >= 0 {
		pkglog.Printf("[ERR]  "+format+"\n", args...)
	}
}

func (l defaultLog) Info(i interface{}) {
	l.Infof("%+v", i)
}

func (l defaultLog) Error(i interface{}) {
	l.Errorf("%+v", i)
}

func (l defaultLog) Close() {
	// Nothing to do
}
