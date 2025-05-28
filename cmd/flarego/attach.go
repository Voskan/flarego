// cmd/flarego/attach.go
// Implements the `flarego attach` command.  For v0.1 this command starts an
// in‑process agent that samples the *current* Go program (i.e., the flarego
// CLI itself) for quick local experimentation.  In later versions it will
// support eBPF‑based dynamic attach to arbitrary PIDs.
//
// Typical usage:
//
//	flarego attach --gateway localhost:4317 --duration 30s
//
// The command spins up a Collector with a GoroutineSampler and a gRPC Exporter
// pointed at the specified gateway address.  It shuts down cleanly on SIGINT
// or after the optional duration elapses.
package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"

	"github.com/Voskan/flarego/internal/agent"
	"github.com/Voskan/flarego/internal/agent/exporter"
	"github.com/Voskan/flarego/internal/agent/sampler"
	"github.com/Voskan/flarego/internal/logging"
)

func newAttachCmd() *cobra.Command {
    var (
        gatewayAddr string
        sampleHz    int
        duration    time.Duration
    )

    cmd := &cobra.Command{
        Use:   "attach",
        Short: "Start a local agent and stream samples to a FlareGo gateway",
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx, cancel := context.WithCancel(cmd.Context())
            if duration > 0 {
                ctx, cancel = context.WithTimeout(ctx, duration)
            }
            defer cancel()

            // Set up collector.
            col := agent.NewCollector(agent.Config{
                Hz:          sampleHz,
                ExportEvery: 500 * time.Millisecond,
            })
            gs := sampler.NewGoroutineSampler(col.Builder(), sampleHz)
            col.AddSampler(gs)

            exp, err := exporter.NewGRPCExporter(ctx, exporter.Config{
                Addr: gatewayAddr,
            })
            if err != nil {
                return err
            }
            col.AddExporter(exp)

            // Start sampling.
            col.Start()
            logging.Sugar().Infow("agent started", "gateway", gatewayAddr, "hz", sampleHz)

            // Handle Ctrl‑C.
            sigCh := make(chan os.Signal, 1)
            signal.Notify(sigCh, os.Interrupt)
            select {
            case <-ctx.Done():
                logging.Sugar().Info("duration elapsed – stopping agent")
            case <-sigCh:
                logging.Sugar().Info("received interrupt – stopping agent")
            }

            col.Stop()
            _ = exp.Close()
            return nil
        },
    }

    cmd.Flags().StringVar(&gatewayAddr, "gateway", "localhost:4317", "FlareGo gateway gRPC address (host:port)")
    cmd.Flags().IntVar(&sampleHz, "hz", 100, "Sampling frequency in Hz (1‑10000)")
    cmd.Flags().DurationVar(&duration, "duration", 0, "Optional run time (e.g., 30s); 0 = run until Ctrl‑C")
    return cmd
}
