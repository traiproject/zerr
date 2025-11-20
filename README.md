# zerr - High-Performance Error Handling Library

zerr is a production-ready, high-performance Go error handling library that
provides modern, idiomatic error wrapping with lazy stack traces, deduplication,
and structured metadata.

## Features

- **Lazy & Deduplicated Stack Traces**: Capture stack traces lazily using
  `runtime.Callers` and deduplicate them using a global cache
- **Efficient & Deduplicated Stack Traces**: Capture stack traces efficiently
  using pooled buffers and deduplicate them using a global cache, deferring
  expensive symbol resolution until formatting
- **Low-Overhead Wrapping**: Optimized happy path with minimal allocation
  overhead for error wrapping
- **Structured Metadata**: Attach typed key-value pairs to errors efficiently
- **Native `slog` Integration**: Automatic structured logging with
  `slog.LogValuer` implementation
- **Goroutine Safety**: Safe recovery from panics in goroutines with `Defer()`
- **Typed Nil Issue Fix**: Fixed the common Go bug where nil pointers with types aren't truly nil

## Requirements

- Go 1.25 or later

## Installation

```bash
go get go.trai.ch/zerr
```

## Usage

> **API Note (v0.2+)**: The `New` and `Wrap` functions now return `error` instead of `*Error` to fix the [typed nil issue](#typed-nil-issue-fix). To use methods like `With` and `WithStack`, you must cast the result to `*zerr.Error`.

### Basic Error Creation

```go
import "go.trai.ch/zerr

// Create a new error
err := zerr.New("something went wrong")

// Wrap an existing error
err = zerr.Wrap(err, "failed to process request")
```

### Adding Metadata

```go
// Add structured metadata
err := zerr.New("database error")
if zerrErr, ok := err.(*zerr.Error); ok {
    err = zerrErr.With("table", "users").
        With("operation", "insert").
        With("user_id", 12345)
}
```

### Stack Traces

```go
// Capture stack trace lazily
err := zerr.New("critical failure")
if zerrErr, ok := err.(*zerr.Error); ok {
    err = zerrErr.WithStack()
}

// Stack traces are deduplicated and cached
err1 := zerr.New("error")
if zerrErr, ok := err1.(*zerr.Error); ok {
    err1 = zerrErr.WithStack()
}
err2 := zerr.New("another error")
if zerrErr, ok := err2.(*zerr.Error); ok {
    err2 = zerrErr.WithStack() // Same stack trace, reused
}
```

### Logging with slog

```go
import "log/slog"

// Log errors with structured fields
zerr.Log(context.Background(), slog.Default(), err)

// Errors automatically format themselves when logged
logger.Error("operation failed", "error", err)
```

### Goroutine Safety

```go
func backgroundTask() {
    defer zerr.Defer(func(err error) {
        // Handle recovered errors
        zerr.Log(context.Background(), slog.Default(), err)
    })

    // Potentially panicking code
    panic("something went wrong")
}
```

## Performance

Benchmarks run on Apple M4 Pro (Go 1.25) demonstrate the efficiency of the
deduplication engine:

```text
BenchmarkNew-14                 1000000000               0.2295 ns/op          0 B/op          0 allocs/op
BenchmarkWrap-14                1000000000               0.2239 ns/op          0 B/op          0 allocs/op
BenchmarkWrapWithStack-14        8360919               142.9 ns/op             0 B/op          0 allocs/op
BenchmarkWithMetadata-14        15398614                77.54 ns/op          200 B/op          4 allocs/op
BenchmarkErrorFormatting-14     1000000000               0.9020 ns/op          0 B/op          0 allocs/op
```

Note: BenchmarkWrapWithStack achieving 0 allocations demonstrates the
effectiveness of the global stack cache. Once a specific stack trace is
captured, subsequent errors from the same location incur no memory allocation
overhead.

The happy path (New/Wrap) is highly optimized. Wrap incurs minimal overhead, and
heavy operations like stack tracing use internal pooling and deduplication to
eliminate GC pressure in hot paths.

## License

MIT

```
```
