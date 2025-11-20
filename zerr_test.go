package zerr

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	err := New("test error")
	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got '%s'", err.Error())
	}
}

func TestWrap(t *testing.T) {
	cause := errors.New("cause")
	err := Wrap(cause, "wrapper")
	if err.Error() != "wrapper: cause" {
		t.Errorf("Expected 'wrapper: cause', got '%s'", err.Error())
	}
}

func TestWrapNil(t *testing.T) {
	err := Wrap(nil, "wrapper")
	if err != nil {
		t.Errorf("Expected nil, got '%v'", err)
	}
}

func TestWith(t *testing.T) {
	testErr := New("test")
	err, ok := testErr.(*Error)
	if !ok {
		t.Fatalf("Expected *Error type, got %T", testErr)
	}
	withErr := err.With("key", "value")
	if len(withErr.metadata) != 1 {
		t.Errorf("Expected 1 metadata item, got %d", len(withErr.metadata))
	}
	if withErr.metadata[0].key.Value() != "key" || withErr.metadata[0].value != "value" {
		t.Errorf("Unexpected metadata: %v", withErr.metadata[0])
	}
}

func TestWithStack(t *testing.T) {
	testErr := New("test")
	err, ok := testErr.(*Error)
	if !ok {
		t.Fatalf("Expected *Error type, got %T", testErr)
	}
	withStackErr := err.WithStack()
	if withStackErr.stack == nil {
		t.Error("Expected stack trace to be captured")
	}
}

func TestUnwrap(t *testing.T) {
	cause := errors.New("cause")
	wrappedErr := Wrap(cause, "wrapper")
	err, ok := wrappedErr.(*Error)
	if !ok {
		t.Fatalf("Expected *Error type, got %T", wrappedErr)
	}
	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Error("Unwrap did not return the cause")
	}
}

func TestLogValuer(t *testing.T) {
	// Create an error with metadata, cause, and wrapping
	cause := errors.New("db connection failed")
	wrappedErr := Wrap(cause, "failed to create user")
	err, ok := wrappedErr.(*Error)
	if !ok {
		t.Fatalf("Expected *Error type, got %T", wrappedErr)
	}
	withErr := err.With("user_id", 101)
	// withErr is already *Error, no need to cast
	err = withErr.With("role", "admin")

	// Use a buffer and JSON handler to capture the log output
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	// Log the error using standard slog
	logger.Error("request failed", "error", err)

	// Verify the output contains all parts of the error
	output := buf.String()

	// Check for the error message
	if !strings.Contains(output, `"msg":"failed to create user"`) {
		t.Error("Log output missing error message")
	}

	// Check for the metadata
	if !strings.Contains(output, `"user_id":101`) {
		t.Error("Log output missing user_id metadata")
	}
	if !strings.Contains(output, `"role":"admin"`) {
		t.Error("Log output missing role metadata")
	}

	// Check for the cause
	if !strings.Contains(output, `"cause":"db connection failed"`) {
		t.Error("Log output missing error cause")
	}
}

func TestLogHelper(t *testing.T) {
	// Test the zerr.Log helper function specifically
	testErr := New("timeout")
	err, ok := testErr.(*Error)
	if !ok {
		t.Fatalf("Expected *Error type, got %T", testErr)
	}
	err = err.With("service", "billing")

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	// Call zerr.Log
	Log(context.Background(), logger, err)

	output := buf.String()
	if !strings.Contains(output, `"service":"billing"`) {
		t.Error("zerr.Log helper failed to log metadata")
	}
	if !strings.Contains(output, `"msg":"timeout"`) {
		t.Error("zerr.Log helper failed to log message")
	}
}

func TestDefer(t *testing.T) {
	var capturedErr error
	func() {
		defer Defer(func(err error) {
			capturedErr = err
		})
		panic("test panic")
	}()

	if capturedErr == nil {
		t.Error("Expected error to be captured")
	}

	if capturedErr.Error() != "test panic" {
		t.Errorf("Expected 'test panic', got '%s'", capturedErr.Error())
	}
}

// Keep Examples...
func ExampleNew() {
	err := New("something went wrong")
	fmt.Println(err.Error())
	// Output: something went wrong
}

func ExampleWrap() {
	cause := errors.New("network timeout")
	err := Wrap(cause, "failed to fetch data")
	fmt.Println(err.Error())
	// Output: failed to fetch data: network timeout
}

func ExampleWith() {
	testErr := New("database error")
	err, _ := testErr.(*Error)
	err = err.With("table", "users").With("operation", "insert")
	fmt.Println(err.Error())
	// Output: database error
}
