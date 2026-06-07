package logging

import (
	"go.uber.org/zap"
)

// Common JSON field keys in structured logs.
// The HTTP correlation header follows pkg/constants.RequestIDHeader ("X-Request-ID");
// in logs snake_case stable request_id is used.
const (
	FieldKeyRequestID = "request_id"
)

// WithField creates a generic structured field (delegates to zap.Any).
func WithField(key string, value any) zap.Field {
	return zap.Any(key, value)
}

// WithError creates a field for errors. If err is nil, returns zap.Skip() (does not add key).
func WithError(err error) zap.Field {
	if err == nil {
		return zap.Skip()
	}
	return zap.Error(err)
}
