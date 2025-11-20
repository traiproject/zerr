// Package zerr provides utilities for safe error handling in goroutines.
package zerr

import (
	"fmt"
)

// Defer recovers from panics in goroutines and converts them to errors.
func Defer(handler func(error)) {
	if r := recover(); r != nil {
		err := convertPanicToError(r)
		handler(err)
	}
}

// convertPanicToError converts a panic value to a zerr Error.
func convertPanicToError(r any) *Error {
	switch v := r.(type) {
	case *Error:
		return v
	case error:
		return &Error{
			message: "panic recovered",
			cause:   v,
			stack:   getOrCreateStack(3), // Skip Defer, recover, and this function
		}
	case string:
		return &Error{
			message: v,
			stack:   getOrCreateStack(3),
		}
	default:
		return &Error{
			message: "panic recovered",
			cause:   &Error{message: fmt.Sprintf("%v", v)},
			stack:   getOrCreateStack(3),
		}
	}
}
