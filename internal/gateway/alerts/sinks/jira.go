// internal/gateway/alerts/sinks/jira.go
// Jira sink creates or comments on issues in Atlassian Jira whenever a FlareGo
// alert fires.  The implementation talks to the Jira Cloud REST API v3 using
// basic-auth with an API token (email + token) or OAuth bearer.  To avoid
// spamming duplicates, the sink keeps an in‑memory LRU of recently‑created
// issues keyed by alert rule name.
//
// Caveats:
//   - For brevity this sample covers only the “create issue” path; linking to
//     existing open issues or adding comments can be added later.
//   - In HA gateway setups each instance may race to create an issue; users
//     should configure sticky routing or use an external deduplication layer.
package sinks

import (
	"bytes"
	"container/list"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/Voskan/flarego/internal/logging"
	"github.com/Voskan/flarego/internal/util"
	"go.uber.org/zap"
)

// JiraSink posts a new issue for each unique alert rule.
type JiraSink struct {
    BaseURL   string // e.g. https://your-domain.atlassian.net
    Project   string // project key, e.g. FLR
    IssueType string // e.g. "Bug", "Incident"; default "Task"

    Email     string // for basic auth
    APIToken  string // for basic auth (preferred for Cloud)
    Bearer    string // alternative OAuth token

    Timeout    time.Duration // HTTP timeout, default 8 s
    MaxRetries int           // attempts on failure, default 3

    // dedup LRU
    mu  sync.Mutex
    lru *list.List // list of rule names, newest front
    set map[string]*list.Element
    cap int
}

// NewJiraSink builds a sink with 128‑entry dedup cache.
func NewJiraSink(baseURL, project, email, token string) *JiraSink {
    return &JiraSink{
        BaseURL:    baseURL,
        Project:    project,
        IssueType:  "Task",
        Email:      email,
        APIToken:   token,
        Timeout:    8 * time.Second,
        MaxRetries: 3,
        lru:        list.New(),
        set:        make(map[string]*list.Element),
        cap:        128,
    }
}

// Notify implements alerts.Sink.
func (s *JiraSink) Notify(rule, msg string) {
    if s.BaseURL == "" || s.Project == "" {
        logging.Sugar().Warn("jira sink missing BaseURL or Project; skipping")
        return
    }

    s.mu.Lock()
    if el, ok := s.set[rule]; ok {
        // already seen recently; move to front and skip creation
        s.lru.MoveToFront(el)
        s.mu.Unlock()
        return
    }
    s.mu.Unlock()

    go s.createIssue(rule, msg)
}

func (s *JiraSink) createIssue(rule, msg string) {
    payload := map[string]any{
        "fields": map[string]any{
            "project": map[string]string{"key": s.Project},
            "summary": "FlareGo alert – " + rule,
            "description": msg,
            "issuetype": map[string]string{"name": s.IssueType},
        },
    }
    body, _ := json.Marshal(payload)

    client := &http.Client{Timeout: s.Timeout}
    backoff := util.NewBackoff()

    url := s.BaseURL + "/rest/api/3/issue"
    for attempt := 1; attempt <= s.MaxRetries; attempt++ {
        ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
        req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
        req.Header.Set("Content-Type", "application/json")
        if s.Bearer != "" {
            req.Header.Set("Authorization", "Bearer "+s.Bearer)
        } else {
            token := base64.StdEncoding.EncodeToString([]byte(s.Email + ":" + s.APIToken))
            req.Header.Set("Authorization", "Basic "+token)
        }

        resp, err := client.Do(req)
        cancel()
        if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
            _ = resp.Body.Close()
            s.updateLRU(rule)
            return
        }
        if err == nil {
            _ = resp.Body.Close()
        }
        logging.Logger().Warn("jira create issue failed", zap.String("rule", rule), zap.Int("attempt", attempt), zap.Error(err))
        if attempt == s.MaxRetries {
            break
        }
        time.Sleep(backoff.Next())
    }
}

func (s *JiraSink) updateLRU(rule string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    if el, ok := s.set[rule]; ok {
        s.lru.MoveToFront(el)
        return
    }
    el := s.lru.PushFront(rule)
    s.set[rule] = el
    if s.lru.Len() > s.cap {
        tail := s.lru.Back()
        if tail != nil {
            s.lru.Remove(tail)
            delete(s.set, tail.Value.(string))
        }
    }
}
