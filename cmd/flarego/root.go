// cmd/flarego/root.go
// Root command for the `flarego` CLI. It wires common flags, global
// initialisation (logger, config file, colour output) and adds top‑level
// sub‑commands located in sibling files (attach.go, record.go, replay.go,
// version.go).
//
// Build‑tag `cli` allows excluding the CLI from tiny agent-only builds.
//go:build cli
// +build cli

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/Voskan/flarego/internal/logging"
	"github.com/Voskan/flarego/pkg/flamegraph"
	"github.com/Voskan/flarego/pkg/version"
)

var (
    cfgFile string
    logJSON bool
    rootCmd = &cobra.Command{
        Use:   "flarego",
        Short: "FlareGo – live scheduler flame visualiser",
        Long:  `FlareGo streams Go runtime scheduler events and renders interactive flame‑graphs in real time.`,
        PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
            // Initialise logger exactly once (idempotent).
            if logging.Initialised() {
                return nil
            }
            return initLogger()
        },
    }
)

func init() {
    cobra.OnInitialize(initConfig)

    // Global flags.
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Path to configuration file (YAML/TOML/JSON)")
    rootCmd.PersistentFlags().BoolVar(&logJSON, "log-json", false, "Enable JSON log output (default is human‑friendly console)")

    // Add sub‑commands (defined in other files).
    rootCmd.AddCommand(newAttachCmd())
    rootCmd.AddCommand(newRecordCmd())
    rootCmd.AddCommand(newReplayCmd())
    rootCmd.AddCommand(newVersionCmd())
}

// Execute is called by main.main().
func Execute() {
    if err := rootCmd.Execute(); err != nil {
        _, _ = fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        // Default search: $HOME/.config/flarego/config.{yaml,toml,json}
        home, err := os.UserHomeDir()
        if err == nil {
            viper.AddConfigPath(filepath.Join(home, ".config", "flarego"))
        }
        viper.SetConfigName("config")
    }

    viper.SetEnvPrefix("FLAREGO")
    viper.AutomaticEnv() // read in environment variables that match

    // Load config file if present.
    if err := viper.ReadInConfig(); err == nil {
        logging.Sugar().Infof("Using config file: %s", viper.ConfigFileUsed())
    }
}

func initLogger() error {
    cfg := zap.NewProductionConfig()
    if !logJSON {
        cfg = zap.NewDevelopmentConfig()
    }
    // Add timestamp in RFC3339 for easy copy‑paste.
    cfg.EncoderConfig.EncodeTime = zap.TimeEncoder(func(t time.Time, enc zap.PrimitiveArrayEncoder) {
        enc.AppendString(t.Format(time.RFC3339))
    })

    logger, err := cfg.Build()
    if err != nil {
        return err
    }
    logging.Set(logger)
    logging.Sugar().Infow("FlareGo starting", "go_version", runtime.Version(), "version", version.String())
    return nil
}

// Utility -----------------------------------------------------------------------------------

// mustDecodeFlame loads a flamegraph JSON file.
func mustDecodeFlame(path string) *flamegraph.Frame {
    f, err := os.ReadFile(path)
    if err != nil {
        logging.Sugar().Fatalw("read file", "err", err)
    }
    var root flamegraph.Frame
    if err := root.UnmarshalJSON(f); err != nil {
        logging.Sugar().Fatalw("decode", "err", err)
    }
    return &root
}
