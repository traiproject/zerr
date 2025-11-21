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
- **Typed Nil Issue Fix**: Fixed the common Go bug where nil pointers with types
  aren't truly nil

## Requirements

- Go 1.25 or later

## Installation

```bash
go get go.trai.ch/zerr
```

## Usage

> **API Note**: The `New` and `Wrap` functions return `error` to prevent the
> [typed nil issue](https://www.google.com/search?q=typed-nil-issue-fix). You
> can use the global helper functions `zerr.With` and `zerr.Stack` to add
> context to any error without manual type assertion.

### Basic Error Creation

```go
import "go.trai.ch/zerr"

// Create a new error
err := zerr.New("something went wrong")

// Wrap an existing error
err = zerr.Wrap(err, "failed to process request")
```

### Adding Metadata

You can add metadata using the global helper (works with any error) or by method
chaining (requires casting).

```go
// Option 1: Use the global helper (easiest)
// Automatically upgrades standard errors to zerr.Error
err := zerr.New("database error")
err = zerr.With(err, "table", "users")

// Option 2: Method chaining (fastest for multiple fields)
// Requires type assertion since New() returns standard error
if zerrErr, ok := err.(*zerr.Error); ok {
    err = zerrErr.With("operation", "insert").
        With("user_id", 12345)
}
```

### Stack Traces

Capture stack traces easily using the global `Stack` helper.

```go
// Capture stack trace lazily
err := zerr.New("critical failure")
err = zerr.Stack(err)

// Works with standard errors too (upgrades them)
stdErr := errors.New("standard Go error")
err = zerr.Stack(stdErr)
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
deduplication engine.

The single allocation in `New` and `Wrap` ensures type safety (preventing typed
nil bugs), while the stack trace machinery remains zero-allocation for cached
traces.

```text
BenchmarkNew-14                 93238076                12.55 ns/op           64 B/op          1 allocs/op
BenchmarkWrap-14                92826571                12.76 ns/op           64 B/op          1 allocs/op
BenchmarkWrapWithStack-14        7189615               169.5 ns/op           128 B/op          2 allocs/op
BenchmarkWithMetadata-14        15186404                78.50 ns/op          200 B/op          4 allocs/op
BenchmarkErrorFormatting-14     752197629                1.591 ns/op           0 B/op          0 allocs/op
```

Note: `BenchmarkWrapWithStack` incurring only 1 allocation (for the error struct
itself) demonstrates the effectiveness of the global stack cache. Once a
specific stack trace is captured, adding it to an error incurs no _additional_
memory allocation overhead beyond the error wrapper.

The happy path (New/Wrap) is highly optimized. Heavy operations like stack
tracing use internal pooling and deduplication to eliminate GC pressure in hot
paths.

## License

MIT
