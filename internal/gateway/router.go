// internal/gateway/router.go
// Router wires together the gRPC Gateway server (see server.go) and the HTTP
// listener (listener.go) so callers can start/stop both through a single
// façade.  This file mainly exists to decouple the cmd/flarego-gateway entry
// point from low‑level implementation details and to prepare for future
// REST+GraphQL endpoints.
package gateway

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// Router bundles a Server (gRPC) and an optional HTTP listener.
// The zero value is not usable; construct via NewRouter.
//
// Typical usage:
//   r := gateway.NewRouter(gwCfg, httpCfg)
//   r.Start(ctx)  // blocks until ctx cancel or fatal error
//
// Shut‑down order respects dependencies: HTTP → gRPC.

type Router struct {
    gw   *Server
    httpCfg HTTPConfig

    httpSrv *http.Server
    wg      sync.WaitGroup
}

// NewRouter instantiates the underlying Gateway Server and prepares HTTP
// listener parameters. If httpCfg.ListenAddr is empty only gRPC is exposed.
func NewRouter(gwCfg Config, httpCfg HTTPConfig) (*Router, error) {
    gw, err := New(gwCfg)
    if err != nil {
        return nil, err
    }
    return &Router{gw: gw, httpCfg: httpCfg}, nil
}

// Start launches gRPC + HTTP (if enabled) and blocks until ctx cancellation.
func (r *Router) Start(ctx context.Context) error {
    // Start HTTP first so that subs can connect as soon as gRPC pushes data.
    if r.httpCfg.ListenAddr != "" {
        r.httpSrv = r.gw.StartHTTP(r.httpCfg)
    }

    // gRPC gateway.
    r.wg.Add(1)
    var grpcErr error
    go func() {
        defer r.wg.Done()
        if err := r.gw.ListenAndServe(ctx); err != nil {
            grpcErr = err
        }
    }()

    // Wait for ctx cancel.
    <-ctx.Done()

    // Graceful HTTP shutdown.
    if r.httpSrv != nil {
        shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        _ = r.httpSrv.Shutdown(shutCtx)
        cancel()
    }

    // Wait for gRPC goroutine.
    r.wg.Wait()
    return grpcErr
}
