// internal/util/backoff.go
// Lightweight exponential‑with‑jitter back‑off helper.  While some parts of
// FlareGo use the cenkalti/backoff library, tiny components (e.g., WebSocket
// fan‑out or plugin loaders) prefer not to pull that dependency.  This
// implementation is dependency‑free, deterministic under test and good enough
// for retry loops that do not need full back‑off state machines.
package util

import (
	"math/rand"
	"time"
)

// Backoff is a stateful exponential back‑off calculator with full jitter as
// described in the AWS architecture blog:
//   next = rand(0, cap) where cap = min(base*2^attempt, max)
//
// All fields are exported so callers can tweak them; changing fields after the
// first Next() call is safe and affects subsequent calculations.
type Backoff struct {
    // Base is the initial duration multiplied by 2^attempt.  Default 100 ms.
    Base time.Duration
    // Max is the upper bound for the random cap.  Default 30 s.
    Max time.Duration
    // Attempt counts calls to Next() and can be reset manually.
    Attempt int
    // rng source; can be swapped in tests.
    rng *rand.Rand
}

// NewBackoff returns a Backoff with sane defaults.
func NewBackoff() *Backoff {
    return &Backoff{
        Base: 100 * time.Millisecond,
        Max:  30 * time.Second,
        rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
    }
}

// Next returns the next back‑off duration using full jitter.
func (b *Backoff) Next() time.Duration {
    if b.Base <= 0 {
        b.Base = 100 * time.Millisecond
    }
    if b.Max <= 0 {
        b.Max = 30 * time.Second
    }
    capDur := b.Base << b.Attempt // base * 2^attempt
    if capDur > b.Max {
        capDur = b.Max
    }
    dur := time.Duration(b.rng.Int63n(int64(capDur) + 1))
    b.Attempt++
    return dur
}

// Reset sets Attempt to zero so the next Next() returns a duration within
// [0,Base].
func (b *Backoff) Reset() { b.Attempt = 0 }
