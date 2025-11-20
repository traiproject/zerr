// Package zerr provides slog integration for structured error logging.
package zerr

import (
	"context"
	"log/slog"
)

// Log logs an error using the provided slog.Logger with structured fields.
func Log(ctx context.Context, logger *slog.Logger, err error) {
	logger.ErrorContext(ctx, err.Error(), logFields(err)...)
}

// logFields extracts structured fields from an error for logging.
func logFields(err error) []any {
	var fields []any
	const maxDepth = 100 // Safety limit to prevent infinite loops in cyclic error chains
	depth := 0

	// Traverse the error chain
	for err != nil {
		// Guard against infinite loops
		if depth >= maxDepth {
			fields = append(fields, slog.String("zerr.error", "max recursion depth exceeded"))
			break
		}
		depth++

		if zerr, ok := err.(*Error); ok {
			// Add metadata fields
			for _, meta := range zerr.metadata {
				fields = append(fields, slog.Any(meta.key.Value(), meta.value))
			}

			// Add stack trace if available
			if zerr.stack != nil && zerr.stack.formatted != "" {
				fields = append(fields, slog.String("stacktrace", zerr.stack.formatted))
			}
		}

		// Move to the next error in the chain
		err = unwrap(err)
	}

	return fields
}

// unwrap returns the next error in the error chain.
func unwrap(err error) error {
	if u, ok := err.(interface{ Unwrap() error }); ok {
		return u.Unwrap()
	}
	return nil
}

// LogValue implements slog.LogValuer for automatic formatting when logged.
func (e *Error) LogValue() slog.Value {
	// Create attributes for all metadata
	attrs := make([]slog.Attr, 0, len(e.metadata)+2) // +2 for message and cause

	// Add the error message
	attrs = append(attrs, slog.String("msg", e.message))

	// Add metadata
	for _, meta := range e.metadata {
		attrs = append(attrs, slog.Any(meta.key.Value(), meta.value))
	}

	// Add stack trace if present
	if e.stack != nil && e.stack.formatted != "" {
		attrs = append(attrs, slog.String("stacktrace", e.stack.formatted))
	}

	// Add cause if present
	if e.cause != nil {
		attrs = append(attrs, slog.Any("cause", e.cause))
	}

	return slog.GroupValue(attrs...)
}
