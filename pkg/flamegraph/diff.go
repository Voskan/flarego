// pkg/flamegraph/diff.go
// Diff utilities for flame graphs.  Given two Frame trees – "base" and
// "head" – the algorithm produces a third tree whose node values are the
// difference (head.Value - base.Value) and whose children contain only nodes
// that changed.  This allows the UI to render before/after comparisons with
// colour‑coding: positive (growth) vs negative (shrink).
package flamegraph

// Diff computes the difference between head and base flame graphs.  The
// returned *Frame has Value = head.Value - base.Value.  Children present in
// either tree are diffed recursively; unchanged subtrees (delta == 0 and no
// changed descendants) are pruned so the output is minimal.
func Diff(head, base *Frame) *Frame {
    if head == nil && base == nil {
        return nil
    }
    if head == nil {
        // Treat missing head as zero values.
        head = &Frame{Name: base.Name, Children: map[string]*Frame{}}
    }
    if base == nil {
        base = &Frame{Name: head.Name, Children: map[string]*Frame{}}
    }

    // Copy node name from head if available, otherwise from base.
    node := &Frame{
        Name:     head.Name,
        Children: make(map[string]*Frame),
    }
    node.Value = head.Value - base.Value

    // Union of child keys.
    keys := make(map[string]struct{})
    for k := range head.Children {
        keys[k] = struct{}{}
    }
    for k := range base.Children {
        keys[k] = struct{}{}
    }

    // Recurse.
    for k := range keys {
        child := Diff(head.Children[k], base.Children[k])
        if child == nil {
            continue
        }
        // Only include non‑zero subtrees.
        if child.Value != 0 || len(child.Children) > 0 {
            node.Children[k] = child
        }
    }

    // Prune leaf if nothing changed.
    if node.Value == 0 && len(node.Children) == 0 {
        return nil
    }
    return node
}
