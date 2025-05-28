// internal/logging/logger.go
// Package logging provides a thin global wrapper around zap.Logger so that
// libraries and generated code inside the FlareGo project can log without
// passing loggers explicitly through every call.
//
// The design is intentionally minimal: a single atomic pointer and helper
// accessors.  Tests may swap the logger (e.g., to zaptest.Buffer) without data
// races.  Production code sets the logger once during program start (see
// cmd/flarego/root.go or gateway main).
package logging

import (
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

var l atomic.Pointer[zap.Logger]

// Set installs the given zap.Logger as the global logger.
// Calling Set more than once overwrites the previous logger; this is useful in
// tests.  The function never panics on nil input – it silently downgrades to a
// zap.NewNop().
func Set(logger *zap.Logger) {
    if logger == nil {
        logger = zap.NewNop()
    }
    l.Store(logger)
}

// Logger returns the globally registered *zap.Logger.  If none has been set it
// returns zap.NewNop() so that callers can safely continue.
func Logger() *zap.Logger {
    if logger := l.Load(); logger != nil {
        return logger
    }
    // fast path: install nop once to avoid repeated allocs
    nop := zap.NewNop()
    l.Store(nop)
    return nop
}

// Sugar is shorthand for Logger().Sugar().
func Sugar() *zap.SugaredLogger { return Logger().Sugar() }

// Initialised reports whether a non-nop logger has been set.
func Initialised() bool {
    logger := l.Load()
    return logger != nil && logger != zap.NewNop()
}
