// Package logging provides a global facade over go.uber.org/zap for the application.
// Before Init, the logger is a no-op (zap.NewNop) and does not cause panic in Info/Error/etc.
// After Init, it is recommended to call Sync to empty buffers when closing the process.
// Common fields: FieldKeyRequestID, WithRequestID, WithError, WithField (fields.go).
package logging

import (
	"go.uber.org/zap"
)

// default logger is no-op until Init; avoids panic if logged before startup.
var logger *zap.Logger = zap.NewNop()

// Init configures the global logger (JSON in production, readable console in development).
// It can be called more than once to replace the instance (e.g. in tests).
func Init(production bool) error {
	var (
		l   *zap.Logger
		err error
	)
	if production {
		l, err = zap.NewProduction()
	} else {
		l, err = zap.NewDevelopment()
	}
	if err != nil {
		return err
	}
	logger = l
	return nil
}

// Sync empties the logger buffers. Invoke once during graceful shutdown;
// it is common to ignore the error in stderr: _ = logging.Sync().
func Sync() error {
	return logger.Sync()
}

// Info writes an information message.
func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

// Error writes an error message.
func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

// Warn writes a warning message.
func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

// Debug writes a debug message.
func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

// WithRequestID creates a logging field for the request ID (see FieldKeyRequestID).
func WithRequestID(requestID string) zap.Field {
	return zap.String(FieldKeyRequestID, requestID)
}

// Logger returns the global Zap instance (useful for integrations that require *zap.Logger).
func Logger() *zap.Logger {
	return logger
}

// setLogger replaces the global logger directly. Intended for tests only.
func setLogger(l *zap.Logger) {
	logger = l
}
