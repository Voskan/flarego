// internal/gateway/config.go
// Centralised loader for gateway configuration.  It complements the Config
// struct declared in server.go by populating it from (in precedence order):
//  1. Explicit options struct passed by the caller
//  2. Environment variables prefixed with FLAREGO_GW_
//  3. Optional YAML/TOML/JSON config file path
//
// The loader keeps the dependency footprint small by using spf13/viper which
// is already present for the CLI side.  If viper is absent in a custom build
// tag scenario, callers can still manually construct Config.
package gateway

import (
	"crypto/tls"
	"time"

	"github.com/spf13/viper"
)

// DefaultConfig returns production‐ready defaults suitable for local dev.
func DefaultConfig() Config {
    return Config{
        ListenAddr:   ":4317",
        TLSConfig:    nil, // plaintext by default; enable via config
        AuthToken:    "",
        RetentionDur: 15 * time.Minute,
        MaxClients:   128,
    }
}

// LoadConfig merges file + env into cfg pointer (caller typically passes
// DefaultConfig()).  filePath may be empty.  envPrefix e.g. "FLAREGO_GW".
func LoadConfig(cfg *Config, filePath, envPrefix string) {
    if cfg == nil {
        tmp := DefaultConfig()
        cfg = &tmp
    }

    v := viper.New()
    v.SetEnvPrefix(envPrefix)
    v.AutomaticEnv()

    if filePath != "" {
        v.SetConfigFile(filePath)
        _ = v.ReadInConfig() // treat missing file as non‐fatal
    }

    // Register custom decode hook for TLS config (expects   tls_cert, tls_key)
    v.SetDefault("tls_cert", "")
    v.SetDefault("tls_key", "")

    _ = v.Unmarshal(&cfg)

    certPath := v.GetString("tls_cert")
    keyPath := v.GetString("tls_key")
    if certPath != "" && keyPath != "" {
        if cert, err := tls.LoadX509KeyPair(certPath, keyPath); err == nil {
            cfg.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
        }
    }
}
