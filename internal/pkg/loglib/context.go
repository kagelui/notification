package loglib

import "context"

type contextKey string

var loggerContextKey = contextKey("logger")

// HasLogger checks if logger object is in Context and returns the logger object if it's there
func HasLogger(ctx context.Context) (*Logger, bool) {
	logger, ok := ctx.Value(loggerContextKey).(*Logger)
	return logger, ok
}

// GetLogger returns logger object from Context, else, return default concise logger
func GetLogger(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(loggerContextKey).(*Logger); ok {
		return logger
	}
	return DefaultLogger()
}

// SetLogger sets the logger into the provided context and returns a copy
func SetLogger(ctx context.Context, value *Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, value)
}
