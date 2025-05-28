// pkg/flamegraph/builder.go
// Concurrent builder that aggregates incoming runtime samples into a mutable
// Frame tree.  Sampler goroutines call Add() with stacks + weights, while the
// collector periodically calls Build() to obtain an immutable snapshot for
// export.  Build() employs copy‐on‐write semantics so that sampling continues
// with minimal pause.
package flamegraph

import "sync"

// Sample represents one observation captured by a sampler.
//   • Stack  – root→leaf list of function/pseudo labels (must be non‐empty)
//   • Weight – numeric cost (ns, bytes, count); may be negative for deltas.
type Sample struct {
    Stack  []string
    Weight int64
}

// Builder owns a root Frame and protects it with an RWMutex.  Hot path Add()
// takes only a read lock when descending the tree; per‐node mutexes in Frame
// ensure correctness during concurrent inserts.  Build() takes a write lock
// briefly to swap root pointer and deep copy old tree for the caller.
//
// Usage:
//     b := flamegraph.NewBuilder("root")
//     b.Add(Sample{[]string{"main","doWork"}, 1})
//     snapshot := b.Build() // *Frame safe for concurrent reads

type Builder struct {
    mu   sync.RWMutex
    root *Frame
}

// NewBuilder returns an empty tree with provided root label.
func NewBuilder(rootName string) *Builder {
    return &Builder{root: New(rootName)}
}

// Add merges one sample into the live tree; safe for concurrent use.
func (b *Builder) Add(s Sample) {
    if len(s.Stack) == 0 || s.Weight == 0 {
        return
    }
    b.mu.RLock()
    node := b.root
    b.mu.RUnlock()

    // AddSample handles its own node‐level locking.
    node.AddSample(s.Stack, s.Weight)
}

// Build returns an immutable deep copy of the current tree while resetting the
// internal root to a fresh empty tree so that sampling can continue without
// growing unbounded.
func (b *Builder) Build() *Frame {
    // Swap root under write lock.
    b.mu.Lock()
    oldRoot := b.root
    b.root = New(oldRoot.Name) // preserve root name
    b.mu.Unlock()

    return oldRoot.sortedCopy() // deep copy with deterministic ordering
}


// Reset очищает агрегатор, сохраняя имя корня.
func (b *Builder) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.root = New(b.root.Name)
}
