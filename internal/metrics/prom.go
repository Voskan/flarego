// internal/metrics/prom.go
// Package metrics centralises Prometheus metric registration for all FlareGo
// binaries (agent, gateway).  It exposes typed collectors and helper update
// functions so that code can remain import-cycle‑free.  The package registers
// with the global prometheus.DefaultRegisterer, which callers typically expose
// via the /metrics HTTP handler from the Prometheus client library.
package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
    once sync.Once

    // Gauge metrics ---------------------------------------------------------
    BlockedGoroutines = prometheus.NewGauge(prometheus.GaugeOpts{
        Namespace: "flarego",
        Subsystem: "runtime",
        Name:      "blocked_goroutines",
        Help:      "Number of goroutines currently in a blocked state.",
    })

    HeapBytes = prometheus.NewGauge(prometheus.GaugeOpts{
        Namespace: "flarego",
        Subsystem: "runtime",
        Name:      "heap_bytes",
        Help:      "Current heap size in bytes (runtime.MemStats.Alloc).",
    })

    // Counter metrics -------------------------------------------------------
    GcPauseTotalNs = prometheus.NewCounter(prometheus.CounterOpts{
        Namespace: "flarego",
        Subsystem: "runtime",
        Name:      "gc_pause_total_ns",
        Help:      "Cumulative GC pause time in nanoseconds.",
    })

    ChunksReceivedTotal = prometheus.NewCounter(prometheus.CounterOpts{
        Namespace: "flarego",
        Subsystem: "gateway",
        Name:      "chunks_received_total",
        Help:      "Total number of flamegraph chunks received from agents.",
    })

    Subscribers = prometheus.NewGauge(prometheus.GaugeOpts{
        Namespace: "flarego",
        Subsystem: "gateway",
        Name:      "subscribers",
        Help:      "Current number of active UI subscriber connections.",
    })
)

// Register exports all metrics; safe to call multiple times.
func Register() {
    once.Do(func() {
        prometheus.MustRegister(
            BlockedGoroutines,
            HeapBytes,
            GcPauseTotalNs,
            ChunksReceivedTotal,
            Subscribers,
        )
    })
}

// UpdateRuntimeMetrics updates gauges with latest runtime numbers collected
// elsewhere (typically from agent samplers).
func UpdateRuntimeMetrics(m map[string]int64) {
    if v, ok := m["blocked_goroutines"]; ok {
        BlockedGoroutines.Set(float64(v))
    }
    if v, ok := m["heap_bytes"]; ok {
        HeapBytes.Set(float64(v))
    }
    if v, ok := m["gc_pause_ns"]; ok {
        GcPauseTotalNs.Add(float64(v))
    }
}
