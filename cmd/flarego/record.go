// cmd/flarego/record.go
// Implements the `flarego record` command.  It starts an in‑process agent,
// samples the current program for a fixed duration and writes the aggregated
// flame‑graph to a `.fgo` file on disk.  The output is a gzipped JSON blob so
// that `flarego replay <file>` can load it instantly, while still being
// usable by external tools after decompression.
package main

import (
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/Voskan/flarego/internal/agent"
	"github.com/Voskan/flarego/internal/agent/sampler"
	"github.com/Voskan/flarego/internal/logging"
)

func newRecordCmd() *cobra.Command {
    var (
        outFile    string
        duration   time.Duration
        sampleHz   int
        noCompress bool
    )

    cmd := &cobra.Command{
        Use:   "record",
        Short: "Record a local flame graph snapshot to a .fgo file",
        Long:  `Starts a lightweight agent inside the flarego process, samples runtime activity for the specified duration and stores the resulting flame‑graph JSON (optionally gzipped) to disk.`,
        RunE: func(cmd *cobra.Command, args []string) error {
            if duration <= 0 {
                return fmt.Errorf("--duration must be > 0")
            }
            // Default output filename: flare-2025‑05‑27T18‑00‑00.fgo
            if outFile == "" {
                ts := time.Now().Format("20060102T150405")
                outFile = fmt.Sprintf("flare-%s.fgo", ts)
            }
            if filepath.Ext(outFile) == "" {
                outFile += ".fgo"
            }

            ctx, cancel := context.WithTimeout(cmd.Context(), duration)
            defer cancel()

            // Collector with Goroutine & GC samplers.
            col := agent.NewCollector(agent.Config{
                Hz:          sampleHz,
                ExportEvery: 0, // no live export
            })
            col.AddSampler(sampler.NewGoroutineSampler(col.Builder(), sampleHz))
            col.AddSampler(sampler.NewGCSampler(col.Builder(), 10))
            col.Start()
            logging.Sugar().Infow("recording started", "duration", duration, "hz", sampleHz)

            <-ctx.Done()
            col.Stop()
            root := col.Builder().Build()

            data, err := root.ToJSON()
            if err != nil {
                return err
            }

            f, err := os.Create(outFile)
            if err != nil {
                return err
            }
            defer f.Close()

            if noCompress {
                if _, err := f.Write(data); err != nil {
                    return err
                }
            } else {
                gw := gzip.NewWriter(f)
                if _, err := gw.Write(data); err != nil {
                    _ = gw.Close()
                    return err
                }
                if err := gw.Close(); err != nil {
                    return err
                }
            }

            logging.Sugar().Infow("recording saved", "file", outFile, "size", len(data))
            return nil
        },
    }

    cmd.Flags().DurationVarP(&duration, "duration", "d", 30*time.Second, "Recording duration (e.g., 30s, 2m)")
    cmd.Flags().StringVarP(&outFile, "output", "o", "", "Output .fgo file path (default auto‑named)")
    cmd.Flags().IntVar(&sampleHz, "hz", 100, "Sampling frequency in Hz")
    cmd.Flags().BoolVar(&noCompress, "no-compress", false, "Disable gzip compression of output file")
    return cmd
}
