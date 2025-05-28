// cmd/flarego-gateway/config.go
// Helper for parsing CLI flags and env vars into gateway.Config and HTTPConfig
// structures so that main.go stays minimal.  The two configs are returned
// separately because gRPC (binary protocol) and HTTP (JSON / WebSocket /
// Prometheus) may be served on different addresses.
//
// Environment variables (prefixed FLAREGO_GW_):
//
//	LISTEN        – gRPC listen address (default :4317)
//	HTTP_LISTEN   – HTTP listen address (default :8080)
//	RETENTION     – retention window (e.g., 15m)
//	AUTH_TOKEN    – static bearer token (optional)
//	TLS_CERT      – path to TLS certificate (PEM)
//	TLS_KEY       – path to TLS key (PEM)
//
// Usage pattern from main.go:
//
//	gwCfg, httpCfg := loadGatewayConfig()
package main

import (
	"flag"
	"time"

	"github.com/spf13/viper"

	"github.com/Voskan/flarego/internal/gateway"
)

// loadGatewayConfig parses flags and env vars once during program start.
func loadGatewayConfig() (gateway.Config, gateway.HTTPConfig) {
    // ----- defaults --------------------------------------------------------
    gwCfg := gateway.DefaultConfig()
    httpCfg := gateway.HTTPConfig{ListenAddr: ":8080", EnableMetrics: true}

    // ----- viper env -------------------------------------------------------
    v := viper.New()
    v.SetEnvPrefix("FLAREGO_GW")
    v.AutomaticEnv()

    // ----- flags -----------------------------------------------------------
    listen := flag.String("listen", gwCfg.ListenAddr, "gRPC listen address (host:port)")
    httpListen := flag.String("http-listen", httpCfg.ListenAddr, "HTTP listen address (host:port, empty to disable)")
    tlsCert := flag.String("tls-cert", "", "TLS certificate file (PEM)")
    tlsKey := flag.String("tls-key", "", "TLS key file (PEM)")
    authToken := flag.String("auth-token", "", "Static bearer token (optional)")
    retention := flag.Duration("retention", gwCfg.RetentionDur, "Retention window (e.g., 15m)")
    maxClients := flag.Int("max-clients", gwCfg.MaxClients, "Soft limit on WebSocket subscribers")
    disableMetrics := flag.Bool("no-metrics", false, "Disable Prometheus /metrics endpoint")
    flag.Parse()

    // ----- merge precedence: flags > env > defaults ------------------------

    if v := v.GetString("LISTEN"); v != "" {
        gwCfg.ListenAddr = v
    }
    if v := v.GetString("HTTP_LISTEN"); v != "" {
        httpCfg.ListenAddr = v
    }
    if d := v.GetDuration("RETENTION"); d > 0 {
        gwCfg.RetentionDur = d
    }
    if tok := v.GetString("AUTH_TOKEN"); tok != "" {
        gwCfg.AuthToken = tok
    }
    if c := v.GetString("TLS_CERT"); c != "" {
        *tlsCert = c
    }
    if k := v.GetString("TLS_KEY"); k != "" {
        *tlsKey = k
    }

    // ----- apply flags -----------------------------------------------------
    gwCfg.ListenAddr = *listen
    gwCfg.AuthToken = *authToken
    gwCfg.RetentionDur = *retention
    gwCfg.MaxClients = *maxClients
    httpCfg.ListenAddr = *httpListen
    httpCfg.EnableMetrics = !*disableMetrics

    // TLS handled by gateway.LoadConfig, but honour flags here too.
    if *tlsCert != "" && *tlsKey != "" {
        gwCfg.TLSCertPath = *tlsCert // new field added in server.Config for late binding
        gwCfg.TLSKeyPath = *tlsKey
    }

    // sanity clamps
    if gwCfg.RetentionDur < time.Minute {
        gwCfg.RetentionDur = time.Minute
    }

    return gwCfg, httpCfg
}
