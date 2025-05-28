// internal/agent/sampler/blocked.go
// BlockedSampler counts goroutines that are currently in a blocked state
// (waiting on channel, mutex, select, etc.) and adds the count as a pseudo
// sample to the flamegraph builder so the live flame graph shows a “Blocked”
// band representing scheduler contention pressure.
//
// Counting strategy:
//   - Use runtime.GoroutineProfile and inspect StackRecord.Inactive (bool)
//     which indicates goroutine is not currently running.
//   - This heuristic is coarse but fast and avoids parsing goroutine header
//     strings.  For more fidelity we could parse stack traces, but that adds
//     allocations; accept trade‑off for v0.1.
package sampler

import (
	"runtime"
	"time"

	"github.com/Voskan/flarego/pkg/flamegraph"
)

// BlockedSampler periodically samples goroutine states.
type BlockedSampler struct {
    builder *flamegraph.Builder
    hz      int

    quit chan struct{}
    done chan struct{}
}

// NewBlockedSampler constructs a sampler with 5–500 Hz limits.
func NewBlockedSampler(b *flamegraph.Builder, hz int) *BlockedSampler {
    if hz < 5 {
        hz = 5
    }
    if hz > 500 {
        hz = 500
    }
    return &BlockedSampler{
        builder: b,
        hz:      hz,
        quit:    make(chan struct{}),
        done:    make(chan struct{}),
    }
}

// Start launches background loop.
func (s *BlockedSampler) Start() {
    select {
    case <-s.done:
        return
    default:
    }
    go s.loop()
}

func (s *BlockedSampler) loop() {
    defer close(s.done)

    interval := time.Second / time.Duration(s.hz)
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    buf := make([]runtime.StackRecord, 256)
    for {
        select {
        case <-ticker.C:
            // Capture snapshot.
            for {
                n, _ := runtime.GoroutineProfile(buf)
                if n < len(buf) {
                    buf = buf[:n]
                    break
                }
                buf = make([]runtime.StackRecord, n*2)
            }
            // Count inactive records.
            var blocked int64
            for _, rec := range buf {
                if rec.Stack() == nil {
                    continue // protective, though should not happen
                }
                if rec.Inactive() {
                    blocked++
                }
            }
            if blocked > 0 {
                s.builder.Add(flamegraph.Sample{
                    Stack:  []string{"(Blocked)"},
                    Weight: blocked,
                })
            }
        case <-s.quit:
            return
        }
    }
}

// Stop signals sampler to finish and waits.
func (s *BlockedSampler) Stop() {
    select {
    case <-s.done:
        return
    default:
        close(s.quit)
        <-s.done
    }
}
