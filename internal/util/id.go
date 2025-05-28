// internal/util/id.go
// Universal unique‑ID helper based on ULID (Universally Unique Lexicographically
// Sortable Identifier).  ULIDs are 128‑bit, URL‑safe and preserve chronological
// order which makes them ideal for log correlation and datastore keys.
//
// The implementation exposes two helpers:
//   - New()         – returns a ULID string in canonical Crockford base‑32
//   - MustNew()     – like New but panics on entropy errors (rare)
//
// To avoid excessive syscalls we keep a process‑global monotonic entropy source
// (math/rand wrapped by ulid.Monotonic) seeded from crypto/rand.
package util

import (
	"crypto/rand"
	"encoding/binary"
	"io"
	mrand "math/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

var entropy *ulid.MonotonicEntropy

func init() {
    // Seed math/rand with crypto‑secure random so that ulid monotonic generator
    // starts at an unpredictable state while remaining cheap thereafter.
    var seed int64
    _ = binaryRead(rand.Reader, &seed)
    entropy = ulid.Monotonic(mrand.New(mrand.NewSource(seed)), 0)
}

// New returns a new ULID string or error.
func New() (string, error) {
    id, err := ulid.New(ulid.Timestamp(time.Now()), entropy)
    if err != nil {
        return "", err
    }
    return id.String(), nil
}

// MustNew panics on failure (entropy read errors).
func MustNew() string {
    s, err := New()
    if err != nil {
        panic(err)
    }
    return s
}

// binaryRead is a tiny helper to read crypto/rand into any fixed‑size integer.
func binaryRead(r io.Reader, v interface{}) error {
    return binary.Read(r, binary.BigEndian, v)
}
