// internal/agent/sampler/heap.go
// HeapSampler records the live heap size (runtime.MemStats.Alloc) at a fixed
// interval and feeds it into flamegraph.Builder as a pseudo‑stack ["(Heap)"]
// whose Weight equals the delta in bytes since the previous sample.
//
// This allows the live flame graph to show a growing or shrinking heap band –
// useful for spotting leaks or allocation bursts in real time.
package sampler

import (
	"runtime"
	"time"

	"github.com/Voskan/flarego/pkg/flamegraph"
)

// HeapSampler polls MemStats.Alloc and converts deltas into samples.
type HeapSampler struct {
    builder *flamegraph.Builder
    hz      int

    quit chan struct{}
    done chan struct{}
}

// NewHeapSampler constructs a sampler.  Frequency `hz` is clamped between 1
// and 4 Hz – higher rates rarely add value for heap trends and only waste CPU.
func NewHeapSampler(b *flamegraph.Builder, hz int) *HeapSampler {
    if hz < 1 {
        hz = 1
    }
    if hz > 4 {
        hz = 4
    }
    return &HeapSampler{
        builder: b,
        hz:      hz,
        quit:    make(chan struct{}),
        done:    make(chan struct{}),
    }
}

// Start begins the background loop.  Subsequent calls are no‑ops.
func (s *HeapSampler) Start() {
    select {
    case <-s.done:
        return // already stopped – cannot restart
    default:
    }
    go s.loop()
}

func (s *HeapSampler) loop() {
    defer close(s.done)

    interval := time.Second / time.Duration(s.hz)
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    var prev uint64
    var ms runtime.MemStats
    runtime.ReadMemStats(&ms)
    prev = ms.Alloc

    for {
        select {
        case <-ticker.C:
            runtime.ReadMemStats(&ms)
            cur := ms.Alloc
            var delta int64
            if cur >= prev {
                delta = int64(cur - prev)
            } else {
                // GC freed memory – encode negative delta.
                delta = -int64(prev - cur)
            }
            prev = cur
            if delta == 0 {
                continue
            }
            s.builder.Add(flamegraph.Sample{
                Stack:  []string{"(Heap)"},
                Weight: delta, // bytes signed
            })
        case <-s.quit:
            return
        }
    }
}

// Stop requests goroutine termination and waits.
func (s *HeapSampler) Stop() {
    select {
    case <-s.done:
        return
    default:
        close(s.quit)
        <-s.done
    }
}
