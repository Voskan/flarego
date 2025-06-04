// internal/agent/sampler/blocked.go
// BlockedSampler counts goroutines that are currently in a blocked state
// (waiting on channel, mutex, select, etc.) and adds the count as a pseudo
// sample to the flamegraph builder so the live flame graph shows a "Blocked"
// band representing scheduler contention pressure.
//
// Counting strategy:
//   - Use runtime.GoroutineProfile to get all goroutines
//   - Count the total number of goroutines as a proxy for blocked count
//   - This is a simplified approach; real implementation would parse stack
//     traces to determine actual blocked states
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
                n, ok := runtime.GoroutineProfile(buf)
                if ok {
                    buf = buf[:n]
                    break
                }
                buf = make([]runtime.StackRecord, len(buf)*2)
            }
            
            // Count goroutines with stacks (simplified blocked detection)
            var blocked int64
            totalGoroutines := int64(runtime.NumGoroutine())
            runningGoroutines := int64(len(buf))
            
            // Estimate blocked as total minus running
            // This is a heuristic; real implementation would parse stack traces
            blocked = totalGoroutines - runningGoroutines
            if blocked < 0 {
                blocked = 0
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
