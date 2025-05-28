// internal/agent/collector.go
// Package agent coordinates samplers and exporters within the in‑process
// FlareGo agent.  A Collector owns a shared flamegraph.Builder, fan‑in of one
// or more Samplers, and one or more Exporters that deliver snapshots to the
// gateway.
//
// Typical lifecycle:
//
//	col := agent.NewCollector(agent.Config{Hz: 100, ExportEvery: 500 * time.Millisecond})
//	col.AddSampler(sampler.NewGoroutineSampler(col.Builder(), 100))
//	col.AddExporter(exporter.NewGRPCExporter(ctx, cfg))
//	col.Start()
//	defer col.Stop()
//
// The design keeps application overhead minimal:
//   - Builder.Add is lock‑free outside the Frame‑level mutexes.
//   - Export loop runs in its own goroutine and uses a read‑only copy of the
//     current flamegraph, avoiding contention with samplers.
//   - Samplers may start and stop independently; Collector handles their
//     lifecycle and waits for graceful shutdown.
package agent

import (
	"context"
	"sync"
	"time"

	"github.com/Voskan/flarego/pkg/flamegraph"
)

// Sampler defines the minimal contract any runtime sampler must satisfy.
type Sampler interface {
    Start()
    Stop()
}

// Exporter delivers a flame graph snapshot to an external sink (gateway, file,
// stdout…).  Implementations must be safe for concurrent use.
type Exporter interface {
    Export(ctx context.Context, root *flamegraph.Frame) error
    Close() error
}

// Config tunes the Collector behaviour.
type Config struct {
    // Hz is the default sampling frequency for samplers that honour it; may be
    // overridden by the specific sampler.
    Hz int

    // ExportEvery defines how often the collector snapshots the Builder and
    // ships it to exporters.  Zero disables automatic exporting (caller can
    // invoke TriggerExport manually).
    ExportEvery time.Duration

    // RootName is the display name for the root frame; defaults to "root".
    RootName string
}

// Collector orchestrates sampling and export pipelines.
type Collector struct {
    cfg      Config
    builder  *flamegraph.Builder

    mu        sync.Mutex
    samplers  []Sampler
    exporters []Exporter

    exportT   *time.Ticker
    quit      chan struct{}
    wg        sync.WaitGroup
}

// NewCollector constructs a collector with sensible defaults.
func NewCollector(cfg Config) *Collector {
    if cfg.Hz == 0 {
        cfg.Hz = 1000 // 1 kHz default; individual samplers may downscale.
    }
    if cfg.RootName == "" {
        cfg.RootName = "root"
    }
    return &Collector{
        cfg:     cfg,
        builder: flamegraph.NewBuilder(cfg.RootName),
        quit:    make(chan struct{}),
    }
}

// Builder returns the underlying flamegraph.Builder for direct access (e.g.,
// merging external traces).  Safe for concurrent use.
func (c *Collector) Builder() *flamegraph.Builder { return c.builder }

// AddSampler registers and starts a new sampler under the collector.
// This can be called at any time before or after Start; if Start has already
// been invoked the sampler is started immediately.
func (c *Collector) AddSampler(s Sampler) {
    c.mu.Lock()
    c.samplers = append(c.samplers, s)
    c.mu.Unlock()

    // Start if collector already running.
    select {
    case <-c.quit:
        // collector stopped – do nothing
    default:
        // best‑effort start, sampler ensures idempotence
        s.Start()
    }
}

// AddExporter registers an exporter.  Exporters are expected to be cheap; the
// collector fan‑outs the same snapshot to all of them sequentially.
func (c *Collector) AddExporter(e Exporter) {
    c.mu.Lock()
    c.exporters = append(c.exporters, e)
    c.mu.Unlock()
}

// Start launches all samplers and, if configured, the periodic export loop.
// Calling Start multiple times is safe but only has effect the first time.
func (c *Collector) Start() {
    c.mu.Lock()
    if c.exportT != nil || c.quit == nil {
        c.mu.Unlock()
        return // already running or collector closed
    }

    // Start samplers.
    for _, s := range c.samplers {
        s.Start()
    }

    // Export ticker.
    if c.cfg.ExportEvery > 0 {
        c.exportT = time.NewTicker(c.cfg.ExportEvery)
        c.wg.Add(1)
        go c.runExportLoop()
    }
    c.mu.Unlock()
}

// runExportLoop periodically snapshots the builder and pushes to exporters.
func (c *Collector) runExportLoop() {
    defer c.wg.Done()

    ctx := context.Background()
    for {
        select {
        case <-c.exportT.C:
            c.pushSnapshot(ctx)
        case <-c.quit:
            return
        }
    }
}

// TriggerExport performs an immediate export once; usable even when
// ExportEvery == 0.  Returns the export error of the first failing exporter.
func (c *Collector) TriggerExport(ctx context.Context) error {
    return c.pushSnapshot(ctx)
}

// pushSnapshot grabs a copy of the current flame graph and iterates exporters.
func (c *Collector) pushSnapshot(ctx context.Context) error {
    snapshot := c.builder.Build()

    c.mu.Lock()
    exporters := append([]Exporter(nil), c.exporters...)
    c.mu.Unlock()

    for _, e := range exporters {
        if err := e.Export(ctx, snapshot); err != nil {
            return err
        }
    }
    return nil
}

// Stop gracefully stops export loop, samplers and exporters.
func (c *Collector) Stop() {
    c.mu.Lock()
    if c.quit == nil {
        c.mu.Unlock()
        return // already stopped
    }
    close(c.quit)
    quitCh := c.quit
    c.quit = nil
    t := c.exportT
    c.exportT = nil
    samplers := append([]Sampler(nil), c.samplers...)
    exporters := append([]Exporter(nil), c.exporters...)
    c.mu.Unlock()

    if t != nil {
        t.Stop()
    }

    // Wait for export loop to end.
    c.wg.Wait()

    // Stop samplers concurrently.
    var wg sync.WaitGroup
    for _, s := range samplers {
        wg.Add(1)
        go func(s Sampler) {
            defer wg.Done()
            s.Stop()
        }(s)
    }
    wg.Wait()

    // Close exporters.
    for _, e := range exporters {
        _ = e.Close()
    }

    // ensure we silence any goroutines blocked on quit channel
    close(quitCh)
}
