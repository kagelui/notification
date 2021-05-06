package loglib

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger wraps logrus entry
type Logger struct {
	LogEntry *logrus.Entry
}

// NewLogger returns an instance of Logger based on the provided logrus entry
func NewLogger(logger *logrus.Entry) *Logger {
	return &Logger{LogEntry: logger}
}

// DefaultLogger returns a Logger with app name and env
func DefaultLogger() *Logger {
	return NewLogger(logrus.WithField("app", os.Getenv("APPNAME")).WithField("env", os.Getenv("APPENV")))
}

// WithField adds a single field to the logger's data
func (v *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{LogEntry: v.LogEntry.WithField(key, value)}
}

// WithFields adds a map of fields to the logger's data
func (v *Logger) WithFields(fields map[string]interface{}) *Logger {
	return &Logger{LogEntry: v.LogEntry.WithFields(fields)}
}

// DebugF logs the message using debug level
func (v *Logger) DebugF(message string, args ...interface{}) {
	v.LogEntry.Debugf(message, args...)
}

// InfoF logs the message using info level
func (v *Logger) InfoF(message string, args ...interface{}) {
	v.LogEntry.Infof(message, args...)
}

// WarnF logs the message using warn level
func (v *Logger) WarnF(message string, args ...interface{}) {
	v.LogEntry.Warnf(message, args...)
}

// ErrorF logs the message using error level
func (v *Logger) ErrorF(message string, args ...interface{}) {
	v.LogEntry.Errorf(message, args...)
}

// FatalF logs the message using fatal level
// Exits using os.Exit(1) after logging
func (v *Logger) FatalF(message string, args ...interface{}) {
	v.LogEntry.Fatalf(message, args...)
}

// PanicF logs the message using panic level
// Triggers a panic after logging
func (v *Logger) PanicF(message string, args ...interface{}) {
	v.LogEntry.Panicf(message, args...)
}
