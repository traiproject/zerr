// Package zerr provides a high-performance error handling library with lazy stack traces,
// deduplication, and structured metadata.
package zerr

import (
	"fmt"
	"runtime"
	"sync"
	"unique"
)

// Error represents an error with optional stack trace and metadata.
type Error struct {
	message  string
	cause    error
	stack    *stackCacheEntry
	metadata []metaPair
}

// metaPair holds a key-value pair for metadata.
type metaPair struct {
	key   unique.Handle[string]
	value any
}

// stackCacheEntry holds a cached stack trace.
type stackCacheEntry struct {
	pc        []uintptr
	formatted string
	once      sync.Once
}

// pcPool is a pool of pc slices for reuse.
var pcPool = sync.Pool{
	New: func() any {
		// Start with a reasonable size, will grow if needed
		pcs := make([]uintptr, 128)
		return &pcs
	},
}

// New creates a new error with the given message.
func New(message string) error {
	return &Error{
		message: message,
	}
}

// Wrap wraps an existing error with an additional message.
// If err is nil, Wrap returns nil.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}

	// Try to cast to our error type first
	if zerr, ok := err.(*Error); ok {
		return &Error{
			message: message,
			cause:   zerr,
		}
	}

	return &Error{
		message: message,
		cause:   err,
	}
}

// With attaches a key-value pair to an error.
// If err is already a *Error, it attaches the metadata directly.
// If err is a standard error, it wraps it to allow attaching metadata.
func With(err error, key string, value any) error {
	if err == nil {
		return nil
	}
	if z, ok := err.(*Error); ok {
		return z.With(key, value)
	}
	// Upgrade standard error to zerr.Error safely
	wrapped := Wrap(err, "")
	if z, ok := wrapped.(*Error); ok {
		return z.With(key, value)
	}
	return wrapped
}

// Stack captures the stack trace for the error.
// If err is already a *Error, it attaches the stack trace directly.
// If err is a standard error, it wraps it to capture the stack trace.
func Stack(err error) error {
	if err == nil {
		return nil
	}
	if z, ok := err.(*Error); ok {
		return z.WithStack()
	}
	// Upgrade standard error to zerr.Error safely
	wrapped := Wrap(err, "")
	if z, ok := wrapped.(*Error); ok {
		return z.WithStack()
	}
	return wrapped
}

// With attaches a key-value pair to the error as metadata.
func (e *Error) With(key string, value any) *Error {
	// Create a new error with the additional metadata
	newErr := &Error{
		message:  e.message,
		cause:    e.cause,
		stack:    e.stack,
		metadata: make([]metaPair, len(e.metadata), len(e.metadata)+1),
	}
	copy(newErr.metadata, e.metadata)
	newErr.metadata = append(newErr.metadata, metaPair{
		key:   unique.Make(key),
		value: value,
	})
	return newErr
}

// WithStack captures a stack trace for this error.
func (e *Error) WithStack() *Error {
	entry := getOrCreateStack(2)

	// Return a new error with the stack trace
	return &Error{
		message:  e.message,
		cause:    e.cause,
		stack:    entry,
		metadata: e.metadata,
	}
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.cause == nil {
		return e.message
	}
	// Avoid ": cause" output if message is empty
	if e.message == "" {
		return e.cause.Error()
	}
	return fmt.Sprintf("%s: %s", e.message, e.cause.Error())
}

// Unwrap implements the unwrap interface for error chaining.
func (e *Error) Unwrap() error {
	return e.cause
}

// Format implements the fmt.Formatter interface to allow for printing stack traces.
func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			// Print with stack trace
			fmt.Fprint(s, e.Error())
			if e.stack != nil {
				e.formatStack(s)
			}
			return
		}
		fallthrough
	case 's':
		fmt.Fprint(s, e.Error())
	case 'q':
		fmt.Fprintf(s, "%q", e.Error())
	}
}

// formatStack formats the stack trace for printing.
func (e *Error) formatStack(s fmt.State) {
	if e.stack.formatted != "" {
		fmt.Fprint(s, e.stack.formatted)
		return
	}

	// Format the stack trace
	frames := runtime.CallersFrames(e.stack.pc)
	for {
		frame, more := frames.Next()
		fmt.Fprintf(s, "\n%s:%d %s", frame.File, frame.Line, frame.Function)
		if !more {
			break
		}
	}
}
