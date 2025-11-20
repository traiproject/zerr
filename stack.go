// Package zerr provides stack trace utilities for the error handling library.
package zerr

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"weak"
)

// stackCache stores a map of stack trace hashes to a list of weak pointers.
// Implements separate chaining to handle hash collisions.
var (
	stackCache   = make(map[uintptr][]weak.Pointer[stackCacheEntry])
	stackCacheMu sync.RWMutex
)

// getOrCreateStack captures the current stack trace and returns a cached entry.
// Implements lazy stack trace capture with deduplication using weak references.
func getOrCreateStack(skip int) *stackCacheEntry {
	// Get a pc slice from the pool
	pcsPtr := pcPool.Get().(*[]uintptr)
	pcs := *pcsPtr

	// Reset length to capacity to ensure we can capture the full stack trace
	pcs = pcs[:cap(pcs)]

	// Capture the stack trace
	n := runtime.Callers(skip+1, pcs)

	// If we didn't capture enough, grow the slice
	if n == len(pcs) {
		pcs = make([]uintptr, len(pcs)*2)
		n = runtime.Callers(skip+1, pcs)
	}

	// Trim to actual size
	pcs = pcs[:n]

	// Create a hash of the pcs for caching using a non-commutative algorithm
	var hash uintptr = 17
	for _, pc := range pcs {
		hash = hash*31 + pc
	}

	// ---------------------------------------------------------
	// 1. READ LOCK: Check existing entries (Separate Chaining)
	// ---------------------------------------------------------
	stackCacheMu.RLock()
	entries, ok := stackCache[hash]
	stackCacheMu.RUnlock()

	if ok {
		for _, weakEntry := range entries {
			// Check if the weak pointer is still valid
			if ptr := weakEntry.Value(); ptr != nil {
				// Verify that the cached PCs actually match the current PCs
				if stackMatches(ptr.pc, pcs) {
					// Found it!
					*pcsPtr = pcs
					pcPool.Put(pcsPtr)
					return ptr
				}
			}
		}
	}

	// ---------------------------------------------------------
	// 2. WRITE LOCK: Create and Insert
	// ---------------------------------------------------------
	// Not in cache or found no match in the chain. Create new entry.
	newEntry := &stackCacheEntry{
		pc: make([]uintptr, len(pcs)),
	}
	copy(newEntry.pc, pcs)

	stackCacheMu.Lock()
	// Double-checked locking: Re-read the slice in case another goroutine beat us
	entries, ok = stackCache[hash]

	var foundEntry *stackCacheEntry

	if ok {
		// Re-scan the chain under the write lock.
		// Also clean up nil (garbage collected) entries.
		activeEntries := entries[:0] // Reuse backing array for filtering

		for _, weakEntry := range entries {
			if ptr := weakEntry.Value(); ptr != nil {
				activeEntries = append(activeEntries, weakEntry)
				if foundEntry == nil && stackMatches(ptr.pc, pcs) {
					foundEntry = ptr
				}
			}
		}

		// Update the map with the compacted list (removed dead weak pointers)
		stackCache[hash] = activeEntries
	}

	if foundEntry != nil {
		// Someone else inserted it while we waited for lock
		stackCacheMu.Unlock()

		*pcsPtr = pcs
		pcPool.Put(pcsPtr)

		return foundEntry
	}

	// Append our new entry to the chain (Separate Chaining)
	stackCache[hash] = append(stackCache[hash], weak.Make(newEntry))
	stackCacheMu.Unlock()

	// ---------------------------------------------------------
	// 3. RETURN
	// ---------------------------------------------------------
	*pcsPtr = pcs
	pcPool.Put(pcsPtr)

	return newEntry
}

// stackMatches checks if two PC slices are identical.
func stackMatches(cached []uintptr, current []uintptr) bool {
	if len(cached) != len(current) {
		return false
	}
	for i, pc := range cached {
		if pc != current[i] {
			return false
		}
	}
	return true
}

// formatStackTrace converts a stack trace to a human-readable string.
// Operation is deferred until the stack trace is actually needed.
func formatStackTrace(pc []uintptr) string {
	if len(pc) == 0 {
		return ""
	}

	var sb strings.Builder
	frames := runtime.CallersFrames(pc)

	for {
		frame, more := frames.Next()
		fmt.Fprintf(&sb, "\n%s:%d %s", frame.File, frame.Line, frame.Function)
		if !more {
			break
		}
	}

	return sb.String()
}

// StackTrace returns a formatted stack trace string.
// Uses lazy formatting - the stack trace is only formatted when this method is called.
func (e *Error) StackTrace() string {
	if e.stack == nil {
		return ""
	}

	// Format on first access and cache the result
	e.stack.once.Do(func() {
		e.stack.formatted = formatStackTrace(e.stack.pc)
	})

	return e.stack.formatted
}
