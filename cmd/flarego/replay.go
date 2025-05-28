// cmd/flarego/replay.go
// Implements the `flarego replay` command.  It loads a previously recorded
// `.fgo` file (produced by `flarego record`), decodes the embedded flamegraph
// JSON and provides two output modes:
//  1. Human‑readable summary on stdout (default)
//  2. Full pretty‑printed JSON via `--json`
//
// Future versions will embed a mini HTTP server that renders the same
// WebComponent used by the dashboard, but for v0.1 the focus is on quick CLI
// inspection and piping into other tools.
package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/Voskan/flarego/pkg/flamegraph"
)

func newReplayCmd() *cobra.Command {
    var outputJSON bool

    cmd := &cobra.Command{
        Use:   "replay <file.fgo>",
        Short: "Inspect a recorded .fgo flamegraph file",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            path := args[0]
            f, err := os.Open(path)
            if err != nil {
                return err
            }
            defer f.Close()

            var r io.Reader = f
            if isGzip(path) {
                gr, err := gzip.NewReader(f)
                if err != nil {
                    return err
                }
                defer gr.Close()
                r = gr
            }

            var root flamegraph.Frame
            dec := json.NewDecoder(r)
            if err := dec.Decode(&root); err != nil {
                return fmt.Errorf("decode flamegraph: %w", err)
            }

            if outputJSON {
                enc := json.NewEncoder(os.Stdout)
                enc.SetIndent("", "  ")
                return enc.Encode(root)
            }

            // Human summary – compute quick stats.
            rows := root.Flatten()
            var totalSelf, totalCum int64
            for _, row := range rows {
                totalSelf += row.Self
                if row.Depth == 0 {
                    totalCum = row.Cumulative
                }
            }

            fmt.Printf("File: %s\n", path)
            fmt.Printf("Nodes: %d\n", len(rows))
            fmt.Printf("Cumulative time: %s\n", time.Duration(totalCum))
            fmt.Printf("Self time total: %s\n", time.Duration(totalSelf))
            fmt.Println("Top 10 hottest stacks:")
            for i, row := range rows[:min(10, len(rows))] {
                fmt.Printf("%2d. %-50s %12s\n", i+1, row.Name, time.Duration(row.Cumulative))
            }
            return nil
        },
    }

    cmd.Flags().BoolVar(&outputJSON, "json", false, "Output full flamegraph JSON instead of summary")
    return cmd
}

// isGzip infers gzip compression from file extension or magic bytes.
func isGzip(path string) bool {
    if filepath.Ext(path) == ".fgo" {
        // record command always gzips unless --no-compress; rely on extension.
        return true
    }
    // Fallback: peek first two bytes.
    f, err := os.Open(path)
    if err != nil {
        return false
    }
    defer f.Close()
    var magic [2]byte
    if _, err := io.ReadFull(f, magic[:]); err != nil {
        return false
    }
    return magic[0] == 0x1f && magic[1] == 0x8b
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
