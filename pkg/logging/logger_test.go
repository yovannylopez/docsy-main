package logging

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name       string
		production bool
		wantErr    bool
	}{
		{
			name:       "initialization in development mode",
			production: false,
			wantErr:    false,
		},
		{
			name:       "initialization in production mode",
			production: true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Init(tt.production)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, Logger())
			}
		})
	}
}

func TestLoggingLevels(t *testing.T) {
	core, logs := observer.New(zap.DebugLevel)
	testLogger := zap.New(core)
	setLogger(testLogger)

	tests := []struct {
		name    string
		logFunc func()
		level   string
		message string
	}{
		{
			name:    "Info logging",
			logFunc: func() { Info("information message") },
			level:   "info",
			message: "information message",
		},
		{
			name:    "Error logging",
			logFunc: func() { Error("error message") },
			level:   "error",
			message: "error message",
		},
		{
			name:    "Warn logging",
			logFunc: func() { Warn("warning message") },
			level:   "warn",
			message: "warning message",
		},
		{
			name:    "Debug logging",
			logFunc: func() { Debug("debug message") },
			level:   "debug",
			message: "debug message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous logs
			logs.TakeAll()

			// Execute the logging function
			tt.logFunc()

			// Verify the message was logged
			allLogs := logs.All()
			require.Len(t, allLogs, 1)
			assert.Equal(t, tt.message, allLogs[0].Message)
		})
	}
}

func TestWithRequestID(t *testing.T) {
	tests := []struct {
		name      string
		requestID string
	}{
		{
			name:      "valid request ID",
			requestID: "req-12345",
		},
		{
			name:      "empty request ID",
			requestID: "",
		},
		{
			name:      "request ID with special characters",
			requestID: "req-abc-123-!@#",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := WithRequestID(tt.requestID)
			assert.Equal(t, "request_id", field.Key)
			assert.Equal(t, tt.requestID, field.String)
		})
	}
}

func TestLogger(t *testing.T) {
	// Configure a test logger
	err := Init(false)
	require.NoError(t, err)

	// Verify Logger() returns the correct logger
	l := Logger()
	assert.NotNil(t, l)
	assert.IsType(t, (*zap.Logger)(nil), l)
}

func TestWithField(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value any
	}{
		{
			name:  "string field",
			key:   "text",
			value: "hello world",
		},
		{
			name:  "int field",
			key:   "count",
			value: 42,
		},
		{
			name:  "bool field",
			key:   "enabled",
			value: true,
		},
		{
			name:  "float field",
			key:   "ratio",
			value: 3.14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := WithField(tt.key, tt.value)
			assert.Equal(t, tt.key, field.Key)
		})
	}
}

func TestWithError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantKey string
	}{
		{
			name:    "simple error",
			err:     errors.New("test error"),
			wantKey: "error",
		},
		{
			name:    "nil error",
			err:     nil,
			wantKey: "error",
		},
		{
			name:    "error with long message",
			err:     errors.New("this is a very long error message that contains multiple words"),
			wantKey: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := WithError(tt.err)
			assert.Equal(t, tt.wantKey, field.Key)
		})
	}
}

func TestLoggingWithFields(t *testing.T) {
	// Configure a test logger with observer
	core, logs := observer.New(zap.DebugLevel)
	testLogger := zap.New(core)
	setLogger(testLogger)

	t.Run("logging with custom fields", func(t *testing.T) {
		// Clear previous logs
		logs.TakeAll()

		userField := WithField("user_id", "user-123")
		actionField := WithField("action", "login")
		errorField := WithError(errors.New("authentication error"))

		// Log message with request ID
		Info("user attempted to authenticate", userField, actionField, errorField)

		// Verify the log
		allLogs := logs.All()
		require.Len(t, allLogs, 1)
		assert.Equal(t, "user attempted to authenticate", allLogs[0].Message)

		// Verify the fields are present
		contextMap := allLogs[0].ContextMap()
		assert.Equal(t, "user-123", contextMap["user_id"])
		assert.Equal(t, "login", contextMap["action"])
		assert.Equal(t, "authentication error", contextMap["error"])
	})
}

func TestLoggingWithRequestID(t *testing.T) {
	// Configure a test logger with observer
	core, logs := observer.New(zap.DebugLevel)
	testLogger := zap.New(core)
	setLogger(testLogger)

	t.Run("logging with request ID", func(t *testing.T) {
		// Clear previous logs
		logs.TakeAll()

		requestField := WithRequestID("req-abc-123")

		// Log message with request ID
		Info("request processed", requestField)

		// Verify the log
		allLogs := logs.All()
		require.Len(t, allLogs, 1)
		assert.Equal(t, "request processed", allLogs[0].Message)

		// Verify request_id is present
		contextMap := allLogs[0].ContextMap()
		assert.Equal(t, "req-abc-123", contextMap["request_id"])
	})
}

func TestAllLoggingLevels(t *testing.T) {
	// Configure a test logger with observer
	core, logs := observer.New(zap.DebugLevel)
	testLogger := zap.New(core)
	setLogger(testLogger)

	t.Run("all logging levels", func(t *testing.T) {
		// Clear previous logs
		logs.TakeAll()

		// Log messages at all levels
		Debug("debug message")
		Info("info message")
		Warn("warn message")
		Error("error message")

		// Verify all messages were logged
		allLogs := logs.All()
		require.Len(t, allLogs, 4)

		// Verify levels in order
		assert.Equal(t, "debug message", allLogs[0].Message)
		assert.Equal(t, "info message", allLogs[1].Message)
		assert.Equal(t, "warn message", allLogs[2].Message)
		assert.Equal(t, "error message", allLogs[3].Message)
	})
}

func TestConcurrentLogging(t *testing.T) {
	// Configure a test logger with observer
	core, logs := observer.New(zap.DebugLevel)
	testLogger := zap.New(core)
	setLogger(testLogger)

	t.Run("concurrent logging", func(t *testing.T) {
		// Clear previous logs
		logs.TakeAll()

		numGoroutines := 10
		// Channel to synchronize goroutines
		done := make(chan struct{})

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		// Execute multiple logs concurrently
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				Info("concurrent message", WithField("goroutine_id", id))
			}(i)
		}

		go func() {
			wg.Wait()
			close(done)
		}()

		// Wait for all goroutines to finish
		<-done

		// Verify all messages were logged
		allLogs := logs.All()
		assert.Len(t, allLogs, numGoroutines)

		// Verify all messages are correct
		for _, log := range allLogs {
			assert.Equal(t, "concurrent message", log.Message)
		}
	})
}

func TestWithFieldTypes(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value any
	}{
		{"string", "text", "hello world"},
		{"int", "count", 42},
		{"float", "ratio", 3.14},
		{"bool", "enabled", true},
		{"slice", "items", []string{"a", "b"}},
		{"map", "data", map[string]any{"key": "value"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := WithField(tt.key, tt.value)
			assert.Equal(t, tt.key, field.Key)
		})
	}
}

func TestWithErrorTypes(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"simple error", errors.New("test error")},
		{"nil error", nil},
		{"error with long message", errors.New("this is a very long error message that contains multiple words")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := WithError(tt.err)
			assert.Equal(t, "error", field.Key)
		})
	}
}

func TestWithError_NilError(t *testing.T) {
	t.Run("WithError with nil error", func(t *testing.T) {
		field := WithError(nil)
		assert.Equal(t, "error", field.Key)
	})
}

func TestMultipleInit(t *testing.T) {
	t.Run("multiple initialization", func(t *testing.T) {
		// First initialization
		err := Init(false)
		require.NoError(t, err)
		firstLogger := Logger()
		require.NotNil(t, firstLogger)

		// Second initialization (should work)
		err = Init(true)
		require.NoError(t, err)
		secondLogger := Logger()
		require.NotNil(t, secondLogger)

		Info("before actual Init")
	})
}

func BenchmarkLogging(b *testing.B) {
	// Configure logger for benchmark
	err := Init(true)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("Info", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Info("benchmark message")
		}
	})

	b.Run("InfoWithFields", func(b *testing.B) {
		fields := []zap.Field{
			WithField("key1", "value1"),
			WithField("key2", 42),
			WithField("key3", true),
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Info("message with fields", fields...)
		}
	})

	b.Run("WithError", func(b *testing.B) {
		testError := errors.New("test error")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			field := WithError(testError)
			_ = field
		}
	})
}
