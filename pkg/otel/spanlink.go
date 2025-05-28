// pkg/otel/spanlink.go
// Helper utilities that allow FlareGo samplers or exporters to correlate
// goroutine IDs and stack traces with OpenTelemetry spans.  The helpers are
// intentionally *optional* – the rest of the project only imports this package
// when the Go OpenTelemetry SDK is present in the build.  There are **no**
// direct imports to internal packages so that external users can reuse the
// helpers in their own instrumentation layers.
//
// Key ideas:
//   - `StartLinkedSpan` starts a new child span on the provided Tracer while
//     recording the current goroutine ID in a span attribute so that the
//     FlareGo gateway can match it later.
//   - `GoroutineID()` duplicates the simple (but safe) hack used by many –
//     parsing runtime.Stack with a small buffer.  It avoids cgo or unsafe.
//   - `WithGID` sets a baggage item with gid so downstream services can look
//     it up even if the span context is lost.
//
// Consumers (agent side) typically wrap long‐running goroutines they spin up:
//
//	func worker(ctx context.Context) {
//	    ctx, span := spanlink.StartLinkedSpan(ctx, tracer, "worker")
//	    defer span.End()
//	    ...
//	}
package otel

import (
	"context"
	"runtime"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
)

const attrGIDKey = "runtime.gid"

// GoroutineID returns the numeric ID of the current goroutine by parsing the
// stack trace header.  It is cheap (~30 ns) and safe because the header format
// is stable since Go 1.4.
func GoroutineID() uint64 {
    var buf [64]byte
    n := runtime.Stack(buf[:], false)
    // first line looks like: "goroutine 12345 [running]:\n"
    fields := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))
    if len(fields) == 0 {
        return 0
    }
    id, _ := strconv.ParseUint(fields[0], 10, 64)
    return id
}

// StartLinkedSpan starts a child span of the span in ctx (or a root span if ctx
// has none) and attaches the current goroutine ID as an attribute so that
// FlareGo can cross‐reference at the gateway.
func StartLinkedSpan(ctx context.Context, tracer trace.Tracer, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
    gid := GoroutineID()
    attr := attribute.Int64(attrGIDKey, int64(gid))
    opts = append(opts, trace.WithAttributes(attr))
    return tracer.Start(ctx, name, opts...)
}

// WithGID returns a context that carries a baggage item "runtime.gid".
// This is helpful when span context propagation is broken – downstream
// services can still read the goroutine ID and annotate their own spans.
func WithGID(ctx context.Context) context.Context {
    gid := GoroutineID()
    member, _ := baggage.NewMember(attrGIDKey, strconv.FormatUint(gid, 10))
    bg, _ := baggage.FromContext(ctx).SetMember(member)
    return baggage.ContextWithBaggage(ctx, bg)
}
