package zerr

import (
	"errors"
	"testing"
)

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = New("test error")
	}
}

func BenchmarkWrap(b *testing.B) {
	err := errors.New("base error")
	b.ReportAllocs()
	for b.Loop() {
		_ = Wrap(err, "wrapper")
	}
}

func BenchmarkWrapWithStack(b *testing.B) {
	err := errors.New("base error")
	b.ReportAllocs()
	for b.Loop() {
		wrapped := Wrap(err, "wrapper")
		zerr, _ := wrapped.(*Error)
		_ = zerr.WithStack()
	}
}

func BenchmarkWithMetadata(b *testing.B) {
	testErr := New("test error")
	err, _ := testErr.(*Error)
	b.ReportAllocs()
	for b.Loop() {
		_ = err.With("key1", "value1").With("key2", 42)
	}
}

func BenchmarkErrorFormatting(b *testing.B) {
	testErr := New("test error")
	err, _ := testErr.(*Error)
	err = err.WithStack()
	b.ReportAllocs()
	for b.Loop() {
		_ = err.Error()
	}
}
