// internal/agent/sampler/gc.go
// GC sampler records stop‑the‑world (STW) garbage‑collection pauses and feeds
// them into the shared flamegraph.Builder so that GC activity appears as a
// separate band in the live flame graph.  Each GC event is translated into a
// pseudo‑stack ["(GC)"] with the weight equal to the pause duration in
// nanoseconds.
//
// Design notes:
//   - runtime.ReadMemStats gives cumulative pause history (last 256 pauses).
//   - The sampler polls NumGC; when the counter increments we inspect the
//     matching PauseNs slot(s) and emit Sample entries.
//   - Overhead is negligible (<0.1 % CPU) even at 100 Hz because ReadMemStats
//     is cheap.
package sampler

import (
	"runtime"
	"sync/atomic"
	"time"

	"github.com/Voskan/flarego/pkg/flamegraph"
)

// GCSampler watches runtime.MemStats and emits GC pause events.
type GCSampler struct {
    builder *flamegraph.Builder
    hz      int

    quit chan struct{}
    done chan struct{}

    lastGCCount uint32 // accessed atomically
}

// NewGCSampler constructs a sampler with frequency `hz` polls per second.
func NewGCSampler(b *flamegraph.Builder, hz int) *GCSampler {
    if hz < 1 {
        hz = 10 // GC does not need very high granularity
    }
    if hz > 1000 {
        hz = 1000
    }
    return &GCSampler{
        builder: b,
        hz:      hz,
        quit:    make(chan struct{}),
        done:    make(chan struct{}),
    }
}

// Start begins the background goroutine if not already running.
func (s *GCSampler) Start() {
    select {
    case <-s.done:
        // already stopped – cannot restart
        return
    default:
    }

    go s.loop()
}

func (s *GCSampler) loop() {
    defer close(s.done)

    interval := time.Second / time.Duration(s.hz)
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    var stats runtime.MemStats
    runtime.ReadMemStats(&stats)
    atomic.StoreUint32(&s.lastGCCount, stats.NumGC)

    for {
        select {
        case <-ticker.C:
            runtime.ReadMemStats(&stats)
            prev := atomic.LoadUint32(&s.lastGCCount)
            cur := stats.NumGC
            if cur == prev {
                continue // no new GC cycle
            }
            // NumGC wraps every 2^32 GC cycles; handle wrap‑around.
            for i := prev; i != cur; i++ {
                idx := int(i%uint32(len(stats.PauseNs)))
                pause := stats.PauseNs[idx]
                if pause == 0 {
                    continue
                }
                s.builder.Add(flamegraph.Sample{
                    Stack:  []string{"(GC)"},
                    Weight: int64(pause),
                })
            }
            atomic.StoreUint32(&s.lastGCCount, cur)
        case <-s.quit:
            return
        }
    }
}

// Stop requests the sampler to stop and waits for termination.
func (s *GCSampler) Stop() {
    select {
    case <-s.done:
        return // already stopped
    default:
        close(s.quit)
        <-s.done
    }
}
