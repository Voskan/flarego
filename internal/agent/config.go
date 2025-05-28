// internal/agent/config.go
// Centralised configuration loader for the FlareGo *agent* binary / library.
// Consumers (cmd/flarego-agent and embedded SDK users) can either:
//   - Call Load() to read config from environment variables + optional YAML
//     file path, or
//   - Instantiate Config struct manually and pass to agent.NewCollector.
//
// The implementation purposefully avoids taking an external YAML dependency;
// when a config file is specified we rely on github.com/spf13/viper which is
// already a transitive dependency of the CLI layer.  If viper is unavailable
// (e.g., tiny embedded build), the file path is ignored gracefully.
package agent

import (
	"time"

	"github.com/spf13/viper"
)

// Config duplicates fields from internal/agent.Collector Config plus exporter
// and encoder choices.
type Config struct {
    // Sampling -------------------------------------------
    Hz int `mapstructure:"hz"` // default 1000

    // Export ---------------------------------------------
    GatewayAddr string        `mapstructure:"gateway_addr"`
    ExportEvery time.Duration `mapstructure:"export_every"`

    // Encoder format: "json" (default) or "proto".
    Encoder string `mapstructure:"encoder"`

    // File exporter (optional).
    FileDir      string `mapstructure:"file_dir"`
    FileCompress bool   `mapstructure:"file_compress"`
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
    return Config{
        Hz:          1000,
        ExportEvery: 500 * time.Millisecond,
        Encoder:     "json",
        GatewayAddr: "localhost:4317",
    }
}

// Load reads configuration from env + optional file.  envPrefix e.g. "FLAREGO"
// transforms AGENT_HZ → Hz.  If filePath is empty only env vars are used.
func Load(filePath, envPrefix string) Config {
    cfg := DefaultConfig()

    v := viper.New()
    if envPrefix != "" {
        v.SetEnvPrefix(envPrefix)
        v.AutomaticEnv()
    }
    if filePath != "" {
        v.SetConfigFile(filePath)
        _ = v.ReadInConfig() // ignore error; treat as optional
    }
    _ = v.Unmarshal(&cfg) // best‐effort merge env + file → struct
    return cfg
}
