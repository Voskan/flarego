// pkg/version/version.go
// Package version holds build-time metadata for the FlareGo binaries.  Values
// are intended to be injected via -ldflags at compile time, e.g.:
//
//	go build -ldflags "-X 'github.com/Voskan/flarego/pkg/version.version=v0.1.0' \
//	                      -X 'github.com/Voskan/flarego/pkg/version.commit=$(git rev-parse --short HEAD)' \
//	                      -X 'github.com/Voskan/flarego/pkg/version.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" ./cmd/flarego
//
// If any variable is left empty, it falls back to a placeholder so that
// Version() always returns a non‑empty string.
package version

import "fmt"

var (
    version = "dev"
    commit  = "unknown"
    date    = "unknown"
)

// String returns a human‑readable representation suitable for --version
// outputs, HTTP headers, etc.
func String() string {
    return fmt.Sprintf("%s (%s, %s)", version, commit, date)
}

// Components returns individual pieces; useful for structured JSON endpoints.
func Components() (ver, gitCommit, buildDate string) {
    return version, commit, date
}
