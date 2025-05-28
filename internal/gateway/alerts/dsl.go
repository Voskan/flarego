// internal/gateway/alerts/dsl.go
// Package alerts implements a _very_ small expression language used in the
// gateway's alert engine.  Its goal is to evaluate boolean conditions over the
// most recent scheduler metrics (e.g., blocked goroutines, GC pause length)
// with minimal allocations and zero third‑party dependencies.
//
// DSL grammar (EBNF):
//
//	Expr   = Ident Sp? Op Sp? Number .
//	Ident  = letter { letter | '_' } ;
//	Op     = '>' | '>=' | '<' | '<=' | '==' | '!=' ;
//	Number = [0-9]+ ;
//	Sp     = { ' ' | '\t' } ;
//
// Example:
//
//	blocked_goroutines > 150
//	heap_bytes >= 536870912
//
// The parser returns a compiled predicate func(map[string]int64) bool that the
// alert engine calls for each incoming chunk.
package alerts

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Predicate evaluates to true when the alert condition is met.
type Predicate func(sample map[string]int64) bool

var (
    errEmptyExpr   = errors.New("empty expression")
    errInvalidExpr = errors.New("invalid expression")
)

// Compile parses s and returns a Predicate or error.
func Compile(s string) (Predicate, error) {
    s = strings.TrimSpace(s)
    if s == "" {
        return nil, errEmptyExpr
    }

    // Scan identifier.
    i := 0
    for i < len(s) && (unicode.IsLetter(rune(s[i])) || s[i] == '_') {
        i++
    }
    if i == 0 {
        return nil, errInvalidExpr
    }
    ident := strings.TrimSpace(s[:i])
    rest := strings.TrimSpace(s[i:])

    // Parse operator.
    opTable := []string{">=", "<=", "!=", "==", ">", "<"}
    var op string
    for _, candidate := range opTable {
        if strings.HasPrefix(rest, candidate) {
            op = candidate
            rest = strings.TrimSpace(rest[len(candidate):])
            break
        }
    }
    if op == "" {
        return nil, fmt.Errorf("%w: missing operator", errInvalidExpr)
    }

    // Number.
    if rest == "" {
        return nil, fmt.Errorf("%w: missing number", errInvalidExpr)
    }
    num, err := strconv.ParseInt(rest, 10, 64)
    if err != nil {
        return nil, fmt.Errorf("%w: %v", errInvalidExpr, err)
    }

    // Build predicate.
    switch op {
    case ">":
        return func(m map[string]int64) bool { return m[ident] > num }, nil
    case ">=":
        return func(m map[string]int64) bool { return m[ident] >= num }, nil
    case "<":
        return func(m map[string]int64) bool { return m[ident] < num }, nil
    case "<=":
        return func(m map[string]int64) bool { return m[ident] <= num }, nil
    case "==":
        return func(m map[string]int64) bool { return m[ident] == num }, nil
    case "!=":
        return func(m map[string]int64) bool { return m[ident] != num }, nil
    default:
        return nil, errInvalidExpr
    }
}
