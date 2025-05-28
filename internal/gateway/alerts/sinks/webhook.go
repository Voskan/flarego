// internal/gateway/alerts/sinks/webhook.go
// Generic webhook sink: performs an HTTP POST with a small JSON payload every
// time an alert fires.  It is often used to integrate FlareGo alerts with chat
// bots, incident managers (PagerDuty, Opsgenie) or custom automation.
//
// The sink is synchronous and retries on transient failures with a configurable
// back‑off (leveraging internal/util/backoff).  To avoid blocking the alert
// engine, the Notify() method off‑loads network operations to a goroutine.
package sinks

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Voskan/flarego/internal/logging"
	"github.com/Voskan/flarego/internal/util"
	"go.uber.org/zap"
)

// WebhookSink posts {rule:"<rule>", msg:"<msg>", ts:<unix>} JSON to URL.
type WebhookSink struct {
    URL       string
    Timeout   time.Duration // per‑request timeout; default 5 s
    MaxRetries int          // total attempts incl. first; default 5
}

// NewWebhookSink returns a sink with defaults.
func NewWebhookSink(url string) *WebhookSink {
    return &WebhookSink{URL: url, Timeout: 5 * time.Second, MaxRetries: 5}
}

// Notify implements alerts.Sink (see alerts/engine.go).  It spawns a goroutine
// so the caller returns immediately.
func (s *WebhookSink) Notify(ruleName, msg string) {
    if s.URL == "" {
        logging.Sugar().Warn("webhook sink configured without URL")
        return
    }
    go s.doPost(ruleName, msg)
}

func (s *WebhookSink) doPost(rule, msg string) {
    payload := map[string]any{
        "rule": rule,
        "msg":  msg,
        "ts":   time.Now().Unix(),
    }
    body, _ := json.Marshal(payload)

    client := &http.Client{Timeout: s.Timeout}
    backoff := util.NewBackoff()

    for attempt := 1; attempt <= s.MaxRetries; attempt++ {
        ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
        req, _ := http.NewRequestWithContext(ctx, http.MethodPost, s.URL, bytes.NewReader(body))
        req.Header.Set("Content-Type", "application/json")

        resp, err := client.Do(req)
        cancel()
        if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
            _ = resp.Body.Close()
            return
        }
        if err == nil {
            _ = resp.Body.Close()
        }
        logging.Logger().Warn("webhook notify failed", zap.String("rule", rule), zap.Int("attempt", attempt), zap.Error(err))
        if attempt == s.MaxRetries {
            break
        }
        time.Sleep(backoff.Next())
    }
}
