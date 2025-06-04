// cmd/flarego-gateway/main.go
// Binary entrypoint for the standalone FlareGo gateway service.  It exposes a
// gRPC endpoint for agents, keeps a time-bounded retention ring and broadcasts
// chunks to WebSocket subscribers (future).  The process is configured via
// CLI flags or environment variables with sane defaults for local testing.
package main

import (
	"context"
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
	gwCfg, httpCfg := loadGatewayConfig()

	if httpCfg.ListenAddr == "" {
		httpCfg.ListenAddr = ":8080"
	}

	// Logger ----------------------------------------------------------------
	lg, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("zap: %v", err)
	}
	logging.Set(lg)
	defer lg.Sync()

	// TLS -------------------------------------------------------------------
	// var tlsCfg *tls.Config
	// if httpCfg.TLSConfig != nil {
	// 	tlsCfg = httpCfg.TLSConfig
	// }

	// Gateway ---------------------------------------------------------------
	gw, err := gateway.New(gwCfg)
	if err != nil {
		lg.Fatal("gateway init", zap.Error(err))
	}

	httpSrv := gw.StartHTTP(httpCfg)

    lg.Info("HTTP config", zap.Any("httpCfg", httpCfg))

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

	// Start gRPC server (blocking)
	if err := gw.ListenAndServe(ctx); err != nil {
		lg.Fatal("serve", zap.Error(err))
	}

	// Shutdown HTTP server
	if httpSrv != nil {
		ctxTimeout, cancel := context.WithTimeout(context.Background(), 40*time.Second)
		_ = httpSrv.Shutdown(ctxTimeout)
		cancel()
	}

	lg.Info("goodbye")
}
