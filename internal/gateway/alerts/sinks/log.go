// internal/gateway/alerts/sinks/log.go
// Log sink simply prints alert firings to the gateway's structured logger.
// It is handy in development or small setups where Slack/email is overkill.
// The sink is non‑blocking and incurs effectively zero overhead.
package sinks

import (
	"github.com/Voskan/flarego/internal/logging"
	"go.uber.org/zap"
)

// LogSink satisfies alerts.Sink.
// No configuration needed; the global zap.Logger is used.
type LogSink struct{}

// NewLogSink returns a singleton instance.
func NewLogSink() *LogSink { return &LogSink{} }

// Notify logs the alert name and message at WARN level.
func (s *LogSink) Notify(ruleName, msg string) {
    logging.Logger().Warn("alert fired", zap.String("rule", ruleName), zap.String("msg", msg))
}
