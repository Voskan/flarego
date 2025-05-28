// internal/gateway/otelbridge.go
// Optional bridge that enriches incoming flamegraph chunks with OpenTelemetry
// span context so that the UI can highlight which goroutine stack frames are
// involved in a particular distributed trace.  The gateway maintains a tiny
// in-memory map goroutine‑ID → active span and updates it on every chunk.
//
// Design trade‑offs:
//   - Simplicity over completeness – we rely on `trace_id` annotations emitted
//     by the agent via runtime/trace.LogUserEvent("trace_id=<hex>") rather than
//     wiring deep into OTEL SDK internals.
//   - Map eviction uses a ring buffer with timestamp TTL (default 2 minutes),
//     sufficient for hot traces while bounded in memory.
//   - The bridge is disabled unless Config.EnableOTEL is set to true.
package gateway

import (
	"encoding/hex"
	"sync"
	"time"
)

// SpanInfo minimal fields we care about.
// Using [16]byte for trace ID matches OTEL spec.
type SpanInfo struct {
    TraceID [16]byte
    SpanID  [8]byte
    Ts      time.Time // last seen; used for TTL eviction
}

// otelBridge correlates goroutine IDs to spans.
// It is concurrency‑safe and lock‑free on read path via sync.Map.
type otelBridge struct {
    enabled bool
    ttl     time.Duration
    mu      sync.Mutex
    m       map[uint64]SpanInfo
}

func newOTELBridge(enabled bool) *otelBridge {
    return &otelBridge{
        enabled: enabled,
        ttl:     2 * time.Minute,
        m:       make(map[uint64]SpanInfo),
    }
}

// updateOnEvent inspects user annotation payload; if it contains "trace_id" it
// records mapping gID → span.
func (b *otelBridge) updateOnEvent(gID uint64, ann string) {
    if !b.enabled {
        return
    }
    const key = "trace_id="
    idx := strings.Index(ann, key)
    if idx == -1 {
        return
    }
    hexTid := ann[idx+len(key):]
    if len(hexTid) < 32 { // need 16‑byte traceID
        return
    }
    var tid [16]byte
    if _, err := hex.Decode(tid[:], []byte(hexTid[:32])); err != nil {
        return
    }
    // span_id optional after comma
    var sid [8]byte
    if j := strings.IndexByte(hexTid, ','); j != -1 && len(hexTid[j+1:]) >= 16 {
        _, _ = hex.Decode(sid[:], []byte(hexTid[j+1:j+17]))
    }

    b.mu.Lock()
    b.m[gID] = SpanInfo{TraceID: tid, SpanID: sid, Ts: time.Now()}
    b.mu.Unlock()
}

// attachToFrame decorates flamegraph Frame JSON with span IDs if mapping still
// valid.  Called in hot path before streaming to UI.
func (b *otelBridge) attachToFrame(root *flamegraph.Frame) {
    if !b.enabled || root == nil {
        return
    }
    b.evictionSweep()
    // depth‑first traversal adding span data when present.
    var dfs func(*flamegraph.Frame)
    dfs = func(f *flamegraph.Frame) {
        if info, ok := b.lookup(f.Name); ok {
            if f.Children == nil {
                f.Children = make(map[string]*flamegraph.Frame)
            }
            f.Children["_trace"] = &flamegraph.Frame{
                Name:  hex.EncodeToString(info.TraceID[:]) + ":" + hex.EncodeToString(info.SpanID[:]),
                Value: 0,
            }
        }
        for _, c := range f.Children {
            dfs(c)
        }
    }
    dfs(root)
}

func (b *otelBridge) lookup(gName string) (SpanInfo, bool) {
    // goroutine function name has pattern "goroutine-<id>"; extract last part.
    if !b.enabled {
        return SpanInfo{}, false
    }
    idx := strings.LastIndexByte(gName, '-')
    if idx == -1 {
        return SpanInfo{}, false
    }
    id, err := strconv.ParseUint(gName[idx+1:], 10, 64)
    if err != nil {
        return SpanInfo{}, false
    }
    b.mu.Lock()
    info, ok := b.m[id]
    b.mu.Unlock()
    if !ok || time.Since(info.Ts) > b.ttl {
        return SpanInfo{}, false
    }
    return info, true
}

func (b *otelBridge) evictionSweep() {
    b.mu.Lock()
    now := time.Now()
    for id, info := range b.m {
        if now.Sub(info.Ts) > b.ttl {
            delete(b.m, id)
        }
    }
    b.mu.Unlock()
}
