// internal/alertsengine/expr.go
// Package alertsengine provides a **safe** and extremely small arithmetic
// expression evaluator used by both the agent and the gateway for advanced
// metric transformations.  Unlike the DSL in internal/gateway/alerts/dsl.go,
// which handles single comparisons (e.g. blocked > 100), this engine supports
// composite formulae such as:
//
//	(heap_bytes / 1024 / 1024) > 512 && blocked_goroutines > 200
//
// Design goals:
//   - Zero dependencies – uses Go's standard library only.
//   - Guard against panics (divide‑by‑zero) and resource exhaustion (max 256
//     AST nodes).
//   - Numeric values are float64 internally; boolean context treats non‑zero as
//     true, zero as false.
//
// Grammar (EBNF):
//
//	Expr   = Or ;
//	Or     = And { "||" And } ;
//	And    = Cmp { "&&" Cmp } ;
//	Cmp    = Add { ( '>' | '>=' | '<' | '<=' | '==' | '!=' ) Add } ;
//	Add    = Mul { ('+'|'-') Mul } ;
//	Mul    = Unary { ('*'|'/') Unary } ;
//	Unary  = [ '!' | '-' ] Primary ;
//	Primary= Number | Ident | '(' Expr ')' ;
//
// Number literals are decimal; Ident matches [a‑zA‑Z_][a‑zA‑Z0‑9_]*.
package alertsengine

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

//--------------------------------------------------------------------
// Public API
//--------------------------------------------------------------------

// Predicate returns true/false when evaluated against a metric map.
type Predicate func(metrics map[string]float64) bool

var (
    ErrSyntax    = errors.New("alertsengine: syntax error")
    ErrNodeLimit = errors.New("alertsengine: AST too deep")
)

// Compile parses expr and returns a Predicate or error.  The caller may cache
// the predicate for repeated evaluations.
func Compile(expr string) (Predicate, error) {
    p := &parser{s: expr, maxNodes: 256}
    node, err := p.parseExpr()
    if err != nil {
        return nil, err
    }
    if p.pos < len(p.s) {
        return nil, fmt.Errorf("%w at %d: unexpected '%s'", ErrSyntax, p.pos, p.s[p.pos:])
    }
    if p.nodeCount > p.maxNodes {
        return nil, ErrNodeLimit
    }
    return func(m map[string]float64) bool {
        v := node.eval(m)
        return v != 0
    }, nil
}

//--------------------------------------------------------------------
// Lexer utilities (minimal ‑ we operate on string indices)
//--------------------------------------------------------------------

type parser struct {
    s         string
    pos       int
    nodeCount int
    maxNodes  int
}

func (p *parser) skipWS() {
    for p.pos < len(p.s) {
        r, sz := utf8.DecodeRuneInString(p.s[p.pos:])
        if r != ' ' && r != '\t' {
            break
        }
        p.pos += sz
    }
}

func (p *parser) match(tok string) bool {
    p.skipWS()
    if strings.HasPrefix(p.s[p.pos:], tok) {
        p.pos += len(tok)
        return true
    }
    return false
}

//--------------------------------------------------------------------
// AST nodes
//--------------------------------------------------------------------

type node interface{
    eval(map[string]float64) float64
}

type binary struct {
    op   string
    lhs, rhs node
}

type unary struct {
    op string
    child node
}

type lit struct{ v float64 }

type ident struct{ name string }

func (b *binary) eval(m map[string]float64) float64 {
    l := b.lhs.eval(m)
    switch b.op {
    case "+":
        return l + b.rhs.eval(m)
    case "-":
        return l - b.rhs.eval(m)
    case "*":
        return l * b.rhs.eval(m)
    case "/":
        r := b.rhs.eval(m)
        if r == 0 {
            return 0
        }
        return l / r
    case "&&":
        if l != 0 && b.rhs.eval(m) != 0 { return 1 }
        return 0
    case "||":
        if l != 0 || b.rhs.eval(m) != 0 { return 1 }
        return 0
    case "==":
        if l == b.rhs.eval(m) { return 1 }
        return 0
    case "!=":
        if l != b.rhs.eval(m) { return 1 }
        return 0
    case ">":
        if l > b.rhs.eval(m) { return 1 }
        return 0
    case ">=":
        if l >= b.rhs.eval(m) { return 1 }
        return 0
    case "<":
        if l < b.rhs.eval(m) { return 1 }
        return 0
    case "<=":
        if l <= b.rhs.eval(m) { return 1 }
        return 0
    default:
        return 0
    }
}

func (u *unary) eval(m map[string]float64) float64 {
    v := u.child.eval(m)
    switch u.op {
    case "-":
        return -v
    case "!":
        if v == 0 {
            return 1
        }
        return 0
    default:
        return v
    }
}

func (l *lit) eval(_ map[string]float64) float64   { return l.v }
func (id *ident) eval(m map[string]float64) float64 { return m[id.name] }

//--------------------------------------------------------------------
// Recursive‑descent parser
//--------------------------------------------------------------------

func (p *parser) newNode(n node) node {
    p.nodeCount++
    return n
}

func (p *parser) parseExpr() (node, error) { return p.parseOr() }

func (p *parser) parseOr() (node, error) {
    left, err := p.parseAnd()
    if err != nil { return nil, err }
    for {
        if p.match("||") {
            right, err := p.parseAnd()
            if err != nil { return nil, err }
            left = p.newNode(&binary{"||", left, right})
        } else {
            return left, nil
        }
    }
}

func (p *parser) parseAnd() (node, error) {
    left, err := p.parseCmp()
    if err != nil { return nil, err }
    for {
        if p.match("&&") {
            right, err := p.parseCmp()
            if err != nil { return nil, err }
            left = p.newNode(&binary{"&&", left, right})
        } else {
            return left, nil
        }
    }
}

var cmpOps = []string{"<=", ">=", "!=", "==", "<", ">"}

func (p *parser) parseCmp() (node, error) {
    left, err := p.parseAdd()
    if err != nil { return nil, err }
    for _, op := range cmpOps {
        if p.match(op) {
            right, err := p.parseAdd()
            if err != nil { return nil, err }
            left = p.newNode(&binary{op, left, right})
            return left, nil
        }
    }
    return left, nil
}

func (p *parser) parseAdd() (node, error) {
    left, err := p.parseMul()
    if err != nil { return nil, err }
    for {
        if p.match("+") {
            right, err := p.parseMul()
            if err != nil { return nil, err }
            left = p.newNode(&binary{"+", left, right})
        } else if p.match("-") {
            right, err := p.parseMul()
            if err != nil { return nil, err }
            left = p.newNode(&binary{"-", left, right})
        } else {
            return left, nil
        }
    }
}

func (p *parser) parseMul() (node, error) {
    left, err := p.parseUnary()
    if err != nil { return nil, err }
    for {
        if p.match("*") {
            right, err := p.parseUnary()
            if err != nil { return nil, err }
            left = p.newNode(&binary{"*", left, right})
        } else if p.match("/") {
            right, err := p.parseUnary()
            if err != nil { return nil, err }
            left = p.newNode(&binary{"/", left, right})
        } else {
            return left, nil
        }
    }
}

func (p *parser) parseUnary() (node, error) {
    if p.match("!") {
        child, err := p.parseUnary()
        if err != nil { return nil, err }
        return p.newNode(&unary{"!", child}), nil
    }
    if p.match("-") {
        child, err := p.parseUnary()
        if err != nil { return nil, err }
        return p.newNode(&unary{"-", child}), nil
    }
    return p.parsePrimary()
}

func (p *parser) parsePrimary() (node, error) {
    p.skipWS()
    if p.match("(") {
        expr, err := p.parseExpr()
        if err != nil { return nil, err }
        if !p.match(")") {
            return nil, ErrSyntax
        }
        return expr, nil
    }
    // number?
    start := p.pos
    for p.pos < len(p.s) && (p.s[p.pos] >= '0' && p.s[p.pos] <= '9' || p.s[p.pos] == '.') {
        p.pos++
    }
    if p.pos > start {
        v, err := strconv.ParseFloat(p.s[start:p.pos], 64)
        if err != nil { return nil, ErrSyntax }
        return p.newNode(&lit{v}), nil
    }
    // identifier
    start = p.pos
    for p.pos < len(p.s) && (isAlphaNum(p.s[p.pos]) || p.s[p.pos]=='_') {
        p.pos++
    }
    if p.pos == start {
        return nil, ErrSyntax
    }
    id := p.s[start:p.pos]
    return p.newNode(&ident{name: id}), nil
}

func isAlphaNum(b byte) bool {
    return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}
