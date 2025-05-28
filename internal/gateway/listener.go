// internal/gateway/listener.go
// HTTP listener that exposes:
//   - /ws   – WebSocket endpoint streaming flamegraph chunks to UI clients
//   - /metrics – optional Prometheus scrape endpoint
//
// The listener is purposely separate from the gRPC server so that deployments
// can route HTTP and gRPC traffic through different ports or ALBs.
package gateway

import (
	"net/http"
	"time"

	"github.com/Voskan/flarego/internal/metrics"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// HTTPConfig controls listener behaviour.
type HTTPConfig struct {
    ListenAddr    string // e.g., ":8080"
    EnableMetrics bool   // expose /metrics
    ReadTimeout   time.Duration
    WriteTimeout  time.Duration
}

// StartHTTP starts an HTTP server in its own goroutine and returns the server
// instance so the caller may shut it down if needed.
func (s *Server) StartHTTP(cfg HTTPConfig) *http.Server {
    if cfg.ReadTimeout == 0 {
        cfg.ReadTimeout = 5 * time.Second
    }
    if cfg.WriteTimeout == 0 {
        cfg.WriteTimeout = 10 * time.Second
    }
    mux := http.NewServeMux()
    mux.HandleFunc("/ws", s.handleWebSocket)
    if cfg.EnableMetrics {
        metrics.Register()
        mux.Handle("/metrics", promhttp.Handler())
    }

    srv := &http.Server{
        Addr:         cfg.ListenAddr,
        Handler:      mux,
        ReadTimeout:  cfg.ReadTimeout,
        WriteTimeout: cfg.WriteTimeout,
    }

    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            s.Logger().Warn("http listener error", zap.Error(err))
        }
    }()
    s.Logger().Info("HTTP listener started", zap.String("addr", cfg.ListenAddr))
    return srv
}

// --------------------------------------------------------------------------------------------------------------------
// WebSocket streaming
// --------------------------------------------------------------------------------------------------------------------

var wsUpgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        // Allow all origins.  In production, restrict as needed.
        return true
    },
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := wsUpgrader.Upgrade(w, r, nil)
    if err != nil {
        s.Logger().Warn("ws upgrade", zap.Error(err))
        return
    }

    ch, unregister := s.Subscribe()
    metrics.Subscribers.Inc()
    defer func() {
        unregister()
        metrics.Subscribers.Dec()
        _ = conn.Close()
    }()

    // Writer loop.
    for buf := range ch {
        if err := conn.WriteMessage(websocket.BinaryMessage, buf); err != nil {
            s.Logger().Debug("ws write", zap.Error(err))
            return
        }
    }
}
