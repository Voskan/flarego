// pkg/flamegraph/frame.go
// Core immutable‐ish data structure representing one node of a flamegraph. The
// frame tree is a prefix‐tree where each node aggregates a numeric Value (e.g.
// nanoseconds, bytes, count).  Children are stored in a map keyed by function
// name for O(1) updates during sampling.  Public methods are concurrent‐safe
// via an internal mutex so that multiple goroutine samplers can mutate the
// tree without global locks.
package flamegraph

import (
	"encoding/json"
	"sort"
	"sync"
)

// Frame is one node in a flamegraph prefix‐tree.
//
//   • Name   – function or pseudo‐stack label
//   • Value  – aggregated metric (self cost; children cumulative added
//              separately by UI)
//   • Children – map[string]*Frame protected by mu
//
// A per‐node mutex allows fine‐grained concurrency during AddSample.
// Access patterns: many reads + writes from samplers, few snapshots; RWLock on
// Builder guards root replacement while Frame itself uses sync.Mutex.
type Frame struct {
    Name     string             `json:"name"`
    Value    int64              `json:"value"`
    Children map[string]*Frame  `json:"children,omitempty"`

    mu sync.Mutex               `json:"-"` // guard Value + Children
}

// New constructs a leaf with given display name.
func New(name string) *Frame {
    return &Frame{Name: name, Children: make(map[string]*Frame)}
}

// AddSample walks stack root→leaf, incrementing Value of each encountered
// node.  Weight may be negative (diff trees, heap deltas).
func (f *Frame) AddSample(stack []string, weight int64) {
    if len(stack) == 0 {
        return
    }
    node := f
    for _, fn := range stack {
        node.mu.Lock()
        child, ok := node.Children[fn]
        if !ok {
            child = &Frame{Name: fn, Children: make(map[string]*Frame)}
            node.Children[fn] = child
        }
        child.Value += weight
        node.mu.Unlock()
        node = child
    }
}

// Merge recursively adds values from src into dst.  Children absent in dst are
// deep‐cloned to preserve isolation.
func (dst *Frame) Merge(src *Frame) {
    if src == nil {
        return
    }
    dst.mu.Lock()
    dst.Value += src.Value
    dst.mu.Unlock()

    for name, sc := range src.Children {
        dst.mu.Lock()
        dc, ok := dst.Children[name]
        if !ok {
            dc = deepCopy(sc)
            dst.Children[name] = dc
            dst.mu.Unlock()
            continue
        }
        dst.mu.Unlock()
        dc.Merge(sc)
    }
}

// ToJSON marshals Frame (and subtree) sorted by descending Value so that the
// UI can assume stable ordering.
func (f *Frame) ToJSON() ([]byte, error) {
    ordered := f.sortedCopy()
    return json.Marshal(ordered)
}

// UnmarshalJSON populates f; implements json.Unmarshaler so that replay command
// can read saved .fgo files back into Frame.
func (f *Frame) UnmarshalJSON(b []byte) error {
    type alias Frame // avoid recursion
    var tmp alias
    if err := json.Unmarshal(b, &tmp); err != nil {
        return err
    }
    *f = Frame(tmp)
    return nil
}

// Flatten returns slice of rows useful for CLI summaries.
type Row struct {
    Name       string
    Depth      int
    Self       int64
    Cumulative int64
}

// Flatten depth‐first traverses tree accumulating cumulative values.
func (f *Frame) Flatten() []Row {
    var rows []Row
    var dfs func(*Frame, int) int64
    dfs = func(n *Frame, depth int) int64 {
        cum := n.Value
        // sort child names for deterministic output
        names := make([]string, 0, len(n.Children))
        for k := range n.Children { names = append(names, k) }
        sort.Strings(names)
        for _, k := range names {
            cum += dfs(n.Children[k], depth+1)
        }
        rows = append(rows, Row{Name: n.Name, Depth: depth, Self: n.Value, Cumulative: cum})
        return cum
    }
    dfs(f, 0)
    // reverse rows so root first
    for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
        rows[i], rows[j] = rows[j], rows[i]
    }
    return rows
}

//--------------------------------------------------------------------
// helpers
//--------------------------------------------------------------------

func deepCopy(src *Frame) *Frame {
    if src == nil { return nil }
    dst := &Frame{Name: src.Name, Value: src.Value, Children: make(map[string]*Frame, len(src.Children))}
    for k, v := range src.Children {
        dst.Children[k] = deepCopy(v)
    }
    return dst
}

// sortedCopy returns deepcopy of tree where each Children map is replaced by a
// slice sorted by Value descending for deterministic UI traversal.
func (f *Frame) sortedCopy() *Frame {
    out := &Frame{Name: f.Name, Value: f.Value}
    if len(f.Children) == 0 {
        return out
    }
    type kv struct{ k string; v *Frame }
    var arr []kv
    for k, v := range f.Children { arr = append(arr, kv{k,v}) }
    sort.Slice(arr, func(i, j int) bool { return arr[i].v.Value > arr[j].v.Value })
    out.Children = make(map[string]*Frame, len(arr))
    for _, kv := range arr {
        out.Children[kv.k] = kv.v.sortedCopy()
    }
    return out
}
