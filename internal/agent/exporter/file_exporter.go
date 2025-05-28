// internal/agent/exporter/file_exporter.go
// File exporter writes each flamegraph snapshot to a directory on the local
// filesystem.  The filename pattern follows
//
//	<prefix>-20060102T150405.000.json[.gz]
//
// where the timestamp is UTC by default.  Compression can be toggled; this
// exporter is primarily for offline analysis and debugging when a gateway is
// unavailable.
package exporter

import (
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Voskan/flarego/pkg/flamegraph"
)

// FileConfig controls exporter behaviour.
type FileConfig struct {
    Dir        string        // destination directory (created if missing)
    Prefix     string        // filename prefix (default "flare")
    Compress   bool          // gzip output
    Timezone   *time.Location // nil => UTC
    FlushSync  bool          // fsync file after write
    Perm       os.FileMode   // file mode (default 0644)
}

// fileExporter implements agent.Exporter.
type fileExporter struct {
    cfg FileConfig
}

// NewFileExporter validates config and returns exporter.
func NewFileExporter(cfg FileConfig) (*fileExporter, error) {
    if cfg.Dir == "" {
        cfg.Dir = "."
    }
    if cfg.Prefix == "" {
        cfg.Prefix = "flare"
    }
    if cfg.Perm == 0 {
        cfg.Perm = 0o644
    }
    if cfg.Timezone == nil {
        cfg.Timezone = time.UTC
    }
    if err := os.MkdirAll(cfg.Dir, 0o755); err != nil {
        return nil, err
    }
    return &fileExporter{cfg: cfg}, nil
}

// Export writes snapshot to file; blocks until write completes.
func (e *fileExporter) Export(_ context.Context, root *flamegraph.Frame) error {
    if root == nil {
        return nil
    }
    data, err := root.ToJSON()
    if err != nil {
        return err
    }
    ts := time.Now().In(e.cfg.Timezone).Format("20060102T150405.000")
    fname := fmt.Sprintf("%s-%s.json", e.cfg.Prefix, ts)
    if e.cfg.Compress {
        fname += ".gz"
    }
    path := filepath.Join(e.cfg.Dir, fname)

    f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_EXCL, e.cfg.Perm)
    if err != nil {
        return err
    }
    defer f.Close()

    if e.cfg.Compress {
        gw := gzip.NewWriter(f)
        if _, err := gw.Write(data); err != nil {
            _ = gw.Close()
            return err
        }
        if err := gw.Close(); err != nil {
            return err
        }
    } else {
        if _, err := f.Write(data); err != nil {
            return err
        }
    }
    if e.cfg.FlushSync {
        _ = f.Sync()
    }
    return nil
}

// Close is a no-op.
func (e *fileExporter) Close() error { return nil }
