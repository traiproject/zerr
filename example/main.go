// Package main demonstrates the usage of the zerr library
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"go.trai.ch/zerr"
)

func main() {
	// Set up structured logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Basic error creation
	err := zerr.New("something went wrong")
	fmt.Printf("Simple error: %s\n", err)

	// Error wrapping with context
	cause := fmt.Errorf("network timeout")
	err = zerr.Wrap(cause, "failed to fetch data from API")
	fmt.Printf("Wrapped error: %s\n", err)

	// Add metadata
	err = zerr.New("database error")
	if zerrErr, ok := err.(*zerr.Error); ok {
		err = zerrErr.With("table", "users").
			With("operation", "insert").
			With("user_id", 12345)
	}
	fmt.Printf("Error with metadata: %s\n", err)

	// Stack traces
	err = zerr.New("critical failure")
	if zerrErr, ok := err.(*zerr.Error); ok {
		err = zerrErr.WithStack()
	}
	fmt.Printf("Error with stack trace:\n%+v\n", err)

	// Logging with slog
	fmt.Println("Logging with slog:")
	zerr.Log(context.Background(), logger, err)

	// Demonstrate zero-allocation happy path
	fmt.Println("\nDemonstrating performance characteristics:")
	err = zerr.New("test")
	if zerrErr, ok := err.(*zerr.Error); ok {
		err = zerrErr.With("key", "value")
	}
	fmt.Printf("Error: %s\n", err)

	// Unwrap errors
	fmt.Printf("Cause: %v\n", err)
}
