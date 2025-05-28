// cmd/flarego-gateway/main.go
// Binary entrypoint for the standalone FlareGo gateway service.  It exposes a
// gRPC endpoint for agents, keeps a time-bounded retention ring and broadcasts
// chunks to WebSocket subscribers (future).  The process is configured via
// CLI flags or environment variables with sane defaults for local testing.
package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Voskan/flarego/internal/gateway"
	"github.com/Voskan/flarego/internal/logging"
	"go.uber.org/zap"
)

func main() {
    // Flags -----------------------------------------------------------------
    listen := flag.String("listen", ":4317", "TCP address to listen on (host:port)")
    tlsCert := flag.String("tls-cert", "", "TLS certificate file (PEM); if empty, serve plaintext")
    tlsKey := flag.String("tls-key", "", "TLS private key file (PEM)")
    authToken := flag.String("auth-token", "", "Static bearer token required from agents (optional)")
    retention := flag.Duration("retention", 15*time.Minute, "In-memory retention window for replay")
    maxClients := flag.Int("max-clients", 128, "Soft cap on concurrent UI subscriber connections")
    flag.Parse()

    // Logger ----------------------------------------------------------------
    lg, err := zap.NewProduction()
    if err != nil {
        log.Fatalf("zap: %v", err)
    }
    logging.Set(lg)
    defer lg.Sync()

    // TLS -------------------------------------------------------------------
    var tlsCfg *tls.Config
    if *tlsCert != "" && *tlsKey != "" {
        cert, err := tls.LoadX509KeyPair(*tlsCert, *tlsKey)
        if err != nil {
            lg.Fatal("load cert", zap.Error(err))
        }
        tlsCfg = &tls.Config{Certificates: []tls.Certificate{cert}}
    }

    // Gateway ---------------------------------------------------------------
    gw, err := gateway.New(gateway.Config{
        ListenAddr:   *listen,
        TLSConfig:    tlsCfg,
        AuthToken:    *authToken,
        RetentionDur: *retention,
        MaxClients:   *maxClients,
    })
    if err != nil {
        lg.Fatal("gateway init", zap.Error(err))
    }

    // Graceful shutdown -----------------------------------------------------
    ctx, cancel := context.WithCancel(context.Background())
    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
        <-sigCh
        lg.Info("signal received, shutting down")
        cancel()
    }()

    // Optional pprof --------------------------------------------------------
    go func() {
        // Expose pprof on 6060 for debugging; ignore errors.
        _ = http.ListenAndServe("localhost:6060", nil)
    }()

    if err := gw.ListenAndServe(ctx); err != nil {
        lg.Fatal("serve", zap.Error(err))
    }

    lg.Info("goodbye")
}
