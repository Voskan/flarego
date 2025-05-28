// pkg/trace/model.go
// Common runtime‑trace event model shared by FlareGo agent, gateway and any
// external integrations that want to reason about low‑level scheduler data.
//
// The design philosophy:
//   - Keep the struct flat and allocation‑cheap so that millions of events can
//     be handled without GC pressure.
//   - Use uint64 for IDs and nsec timestamps for direct mapping to Go’s
//     runtime/trace numbers.
//   - Provide a small EventType enum that covers scheduler + custom user events.
//
// The package purposefully does **not** include any I/O; reader.go and
// filters.go implement deserialisation and helpers.
package trace

import "time"

// EventType identifies a kind of runtime event.
// The numeric values are compatible with Go runtime/trace constants where
// applicable; custom FlareGo events start from 1000.
type EventType uint16

const (
    // Native runtime events (subset).
    EvGoCreate    EventType = 1  // goroutine creation
    EvGoEnd       EventType = 2  // goroutine finished
    EvGoSched     EventType = 3  // goroutine context‑switch
    EvGoBlocked   EventType = 4  // goroutine blocked on sync
    EvGCStart     EventType = 5  // GC cycle start STW
    EvGCEnd       EventType = 6  // GC cycle done

    // FlareGo‑specific synthetic events.
    EvHeapSample  EventType = 1000 // heap size sampled (Value = bytes)
    EvBlockedCnt  EventType = 1001 // total blocked goroutines (Value = count)
)

// Event is a single trace record.
//
// Fields:
//   Ts       – monotonic nanoseconds since process start (like runtime/trace);
//   G        – goroutine ID (0 if N/A, e.g. GC events);
//   P        – processor ID (‑1 if unknown);
//   Type     – event type constant;
//   Value    – generic payload whose meaning depends on Type (e.g., heap
//              bytes, pause ns, count).  0 when unused.
//   Stack    – program counter slice (root→leaf); may be nil to save space.
//
// No pointers back to parent events are stored to keep struct small and to
// avoid link cycles in JSON.
//
// MarshalJSON provided to emit a compact array representation: [ts, g, p, typ,
// value].  Stack omitted unless non‑nil.
//
type Event struct {
    Ts    uint64    `json:"ts"` // monotonic ns
    G     uint64    `json:"g"`
    P     int32     `json:"p"`
    Type  EventType `json:"type"`
    Value int64     `json:"val,omitempty"`
    Stack []uintptr `json:"stack,omitempty"`
}

// Time converts monotonic timestamp to wall‑clock approx.  Caller must supply
// a base wall time corresponding to t=0 (usually time.Now() at trace start).
func (e Event) Time(base time.Time) time.Time {
    return base.Add(time.Duration(e.Ts))
}
