// internal/gateway/alerts/sinks/slack.go
// Slack sink posts a JSON payload to a Slack Incoming Webhook URL whenever an
// alert fires.  It is intentionally minimal and synchronous; consider wrapping
// in a queue for high‑throughput setups.
package sinks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Voskan/flarego/internal/logging"
	"go.uber.org/zap"
)

// SlackSink implements alerts.Sink for Slack.
//
// Example webhook URL format:
//   https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX
//
// The message field supports Slack's mrkdwn.
type SlackSink struct {
    WebhookURL string
    Username   string // optional
    IconEmoji  string // optional (":fire:")
    Timeout    time.Duration
    httpClient *http.Client
}

// NewSlackSink constructs a sink with default HTTP client (10 s timeout).
func NewSlackSink(webhookURL string) *SlackSink {
    return &SlackSink{
        WebhookURL: webhookURL,
        Timeout:    10 * time.Second,
    }
}

// Notify sends msg to Slack with basic retry (3 attempts, linear backoff).
func (s *SlackSink) Notify(ruleName, msg string) {
    if s.WebhookURL == "" {
        logging.Sugar().Warn("Slack sink configured without webhook URL")
        return
    }

    payload := map[string]interface{}{
        "text":       "*FlareGo alert* — " + msg,
        "username":   s.Username,
        "icon_emoji": s.IconEmoji,
    }
    body, _ := json.Marshal(payload)

    cli := s.httpClient
    if cli == nil {
        cli = &http.Client{Timeout: s.Timeout}
    }

    for attempt := 1; attempt <= 3; attempt++ {
        resp, err := cli.Post(s.WebhookURL, "application/json", bytes.NewReader(body))
        if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
            _ = resp.Body.Close()
            return
        }
        if err == nil {
            _ = resp.Body.Close()
        }
        logging.Logger().Warn("Slack notify failed", zap.String("rule", ruleName), zap.Int("attempt", attempt), zap.Error(err))
        time.Sleep(time.Duration(attempt) * time.Second)
    }
}
