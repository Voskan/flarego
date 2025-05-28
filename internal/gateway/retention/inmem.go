// internal/gateway/retention/inmem.go
// Package retention provides pluggable stores that keep recently‑seen
// flame‑graph chunks for replay and late subscribers.  The in‑memory
// implementation uses a time‑bounded ring buffer with O(1) append and O(n)
// expiry, suitable for a single‑instance gateway.  HA deployments should
// replace this with the Redis store or any distributed cache.
package retention

import (
	"sync"
	"time"
)

// Store is a minimal interface required by the gateway server.
//
// Implementations MUST be safe for concurrent use by multiple goroutines.
type Store interface {
    // Write persists one chunk; implementations may mutate the slice to copy
    // the data, therefore callers should treat b as consumed after the call.
    Write(b []byte) error

    // ReadAll returns the currently retained chunks ordered oldest→newest.  The
    // returned slices MUST be deep copies so the caller cannot mutate internal
    // state.
    ReadAll() [][]byte
}

// inMem is a circular buffer that drops chunks older than retentionDur.
type inMem struct {
    retentionDur time.Duration

    mu     sync.RWMutex
    idx    int          // next write position (mod len(buf))
    buf    [][]byte     // ring buffer of chunks
    tsBuf  []time.Time  // parallel slice of timestamps
    filled bool         // becomes true once the ring has wrapped
}

// NewInMem constructs a retention store keeping data for at least d.
//
//   • If d < 1s it is clamped to 1s.
//   • Buffer capacity is sized to hold (d / 100ms) chunks, assuming the agent
//     exports every 100ms (more than enough for default 500ms export cadence).
func NewInMem(d time.Duration) Store {
    if d < time.Second {
        d = time.Second
    }
    // Capacity heuristic: 10× more slots than seconds retained.
    capSlots := int(d.Seconds()*10) + 1
    return &inMem{
        retentionDur: d,
        buf:          make([][]byte, capSlots),
        tsBuf:        make([]time.Time, capSlots),
    }
}

// Write satisfies Store by copying b and appending it to the ring.
func (r *inMem) Write(b []byte) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    now := time.Now()

    // Copy to detach from caller’s buffer.
    cloned := append([]byte(nil), b...)

    r.buf[r.idx] = cloned
    r.tsBuf[r.idx] = now
    r.idx = (r.idx + 1) % len(r.buf)
    if r.idx == 0 {
        r.filled = true
    }

    // Expire old entries lazily (only current slot).
    if !r.filled {
        return nil // nothing to purge yet
    }
    cutoff := now.Add(-r.retentionDur)
    // Check the slot we just overwrote; if it is still within retention we can
    // fast‑path.
    if r.tsBuf[r.idx].After(cutoff) {
        return nil
    }

    // Full scan to drop old elements (rare, only once per retentionDur).
    for i, ts := range r.tsBuf {
        if ts.Before(cutoff) {
            r.buf[i] = nil
            r.tsBuf[i] = time.Time{}
        }
    }
    return nil
}

// ReadAll returns deep copies ordered from oldest to newest.
func (r *inMem) ReadAll() [][]byte {
    r.mu.RLock()
    defer r.mu.RUnlock()

    var res [][]byte

    // Helper to append clone of slice.
    appendClone := func(b []byte) {
        if b == nil {
            return
        }
        res = append(res, append([]byte(nil), b...))
    }

    if !r.filled {
        // Simple linear output up to idx.
        for i := 0; i < r.idx; i++ {
            appendClone(r.buf[i])
        }
        return res
    }

    // Output from idx to end, then start to idx‑1 to maintain chronological order.
    for i := r.idx; i < len(r.buf); i++ {
        appendClone(r.buf[i])
    }
    for i := 0; i < r.idx; i++ {
        appendClone(r.buf[i])
    }
    return res
}
