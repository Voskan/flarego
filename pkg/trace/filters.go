// pkg/trace/filters.go
// Convenience helpers for slicing / dicing Event slices produced by the
// reader.  These helpers are used by the CLI (`flarego replay` future diff
// sub‑command) and by tests when asserting sampler correctness.
//
// The helpers avoid generics to keep Go 1.24 compatibility.
package trace

import "time"

//--------------------------------------------------------------------
// Basic predicate filters
//--------------------------------------------------------------------

// ByTimeRange returns events whose Ts converted with baseTime fall within
// [from, to).  If from.IsZero() it is treated as -∞; if to.IsZero() as +∞.
func ByTimeRange(ev []Event, baseTime time.Time, from, to time.Time) []Event {
    if from.IsZero() && to.IsZero() {
        return clone(ev)
    }
    var out []Event
    for _, e := range ev {
        t := e.Time(baseTime)
        if !from.IsZero() && t.Before(from) {
            continue
        }
        if !to.IsZero() && !t.Before(to) {
            continue
        }
        out = append(out, e)
    }
    return out
}

// ByGoroutineID filters events for a specific goroutine id; id=0 is a no‑op.
func ByGoroutineID(ev []Event, gid uint64) []Event {
    if gid == 0 {
        return clone(ev)
    }
    var out []Event
    for _, e := range ev {
        if e.G == gid {
            out = append(out, e)
        }
    }
    return out
}

// ByEventTypes keeps only events whose Type is in the allow list.  Empty list
// returns clone(ev).  The list is converted to a map for O(1) lookups.
func ByEventTypes(ev []Event, types ...EventType) []Event {
    if len(types) == 0 {
        return clone(ev)
    }
    allow := make(map[EventType]struct{}, len(types))
    for _, t := range types {
        allow[t] = struct{}{}
    }
    var out []Event
    for _, e := range ev {
        if _, ok := allow[e.Type]; ok {
            out = append(out, e)
        }
    }
    return out
}

//--------------------------------------------------------------------
// Utility helpers
//--------------------------------------------------------------------

// Downsample returns every nth event (n>=2).  n<=1 returns clone(ev).
func Downsample(ev []Event, n int) []Event {
    if n <= 1 {
        return clone(ev)
    }
    out := make([]Event, 0, len(ev)/n+1)
    for i := 0; i < len(ev); i += n {
        out = append(out, ev[i])
    }
    return out
}

// AggregateValueByType sums the Value field for each EventType.  Useful for
// quick counters in CLI.
func AggregateValueByType(ev []Event) map[EventType]int64 {
    m := make(map[EventType]int64)
    for _, e := range ev {
        m[e.Type] += e.Value
    }
    return m
}

//--------------------------------------------------------------------
// internal helpers
//--------------------------------------------------------------------

func clone(src []Event) []Event {
    if len(src) == 0 {
        return nil
    }
    dst := make([]Event, len(src))
    copy(dst, src)
    return dst
}
