// cmd/flarego-agent/main.go
// Minimal standalone agent binary.  It embeds the in-process Collector with
// Goroutine / GC / Heap / Blocked samplers and streams data to the configured
// FlareGo Gateway.  Intended for scenarios where you cannot import the agent
// package into the target process but still want to collect traces (e.g., run
// as a sidecar and attach via pprof HTTP endpoints in future versions).  For
// v0.1 the agent simply samples itself – useful for demo and load testing.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Voskan/flarego/internal/agent"
	"github.com/Voskan/flarego/internal/agent/exporter"
	"github.com/Voskan/flarego/internal/agent/sampler"
	"github.com/Voskan/flarego/internal/logging"
	"go.uber.org/zap"
)

func main() {
    // CLI flags -------------------------------------------------------------
    gatewayAddr := flag.String("gateway", "localhost:4317", "FlareGo gateway gRPC address")
    hz := flag.Int("hz", 100, "Sampling frequency in Hz")
    runFor := flag.Duration("duration", 0, "Optional duration to run; 0 = until signal")
    flag.Parse()

    // Logger ----------------------------------------------------------------
    lg, err := zap.NewProduction()
    if err != nil {
        log.Fatalf("zap init: %v", err)
    }
    logging.Set(lg)
    defer lg.Sync()

    // Collector -------------------------------------------------------------
    col := agent.NewCollector(agent.Config{
        Hz:          *hz,
        ExportEvery: 500 * time.Millisecond,
    })
    col.AddSampler(sampler.NewGoroutineSampler(col.Builder(), *hz))
    col.AddSampler(sampler.NewGCSampler(col.Builder(), 10))
    col.AddSampler(sampler.NewHeapSampler(col.Builder(), 2))
    col.AddSampler(sampler.NewBlockedSampler(col.Builder(), 50))

    exp, err := exporter.NewGRPCExporter(context.Background(), exporter.Config{
        Addr: *gatewayAddr,
    })
    if err != nil {
        lg.Fatal("grpc exporter", zap.Error(err))
    }
    col.AddExporter(exp)
    col.Start()
    lg.Info("flarego-agent started", zap.String("gateway", *gatewayAddr), zap.Int("hz", *hz))

    // Shutdown handling -----------------------------------------------------
    done := make(chan struct{})
    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
        select {
        case <-sigCh:
            lg.Info("signal received, shutting down agent")
        case <-time.After(*runFor):
            if *runFor > 0 {
                lg.Info("duration elapsed, shutting down agent")
            }
        }
        col.Stop()
        _ = exp.Close()
        close(done)
    }()

    <-done
    lg.Info("bye")
}
