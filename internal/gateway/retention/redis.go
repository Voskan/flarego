// internal/gateway/retention/redis.go
// Redis-backed retention store – suitable for HA Gateway deployments where
// multiple instances must share flamegraph chunks. The implementation relies
// on a capped Redis list per namespace ("flarego:chunks") with TTL set to the
// retention duration.  Writes are fire‑and‑forget (LPUSH + EXPIRE) for speed;
// reads perform LRANGE to stream the latest N chunks to the new subscriber.
//
// The design assumes Redis ≥ 5.0.  For clusters, use a client that supports
// routing (go-redis/v9 does).  Error handling is kept lenient: write errors are
// logged and swallowed; read errors result in empty slice.
package retention

import (
	"context"
	"time"

	"github.com/Voskan/flarego/internal/logging"
	"github.com/redis/go-redis/v9"
)

const redisKey = "flarego:chunks"

type redisStore struct {
    cli          *redis.Client
    retentionDur time.Duration
    maxLen       int64 // max list length calculated from retentionDur * writes per second
}

// NewRedis returns a Store backed by Redis.  writesPerSecond is an estimate of
// how many chunks will be pushed; it determines list trimming length.
func NewRedis(cli *redis.Client, retention time.Duration, writesPerSecond int) Store {
    if retention < time.Second {
        retention = time.Second
    }
    if writesPerSecond <= 0 {
        writesPerSecond = 10 // default
    }
    maxLen := int64(retention.Seconds()*float64(writesPerSecond)) + 100 // headroom
    return &redisStore{cli: cli, retentionDur: retention, maxLen: maxLen}
}

// Write appends a chunk to Redis list with expiration.
func (r *redisStore) Write(b []byte) error {
    ctx := context.Background()
    pipe := r.cli.Pipeline()
    pipe.LPush(ctx, redisKey, b)
    pipe.LTrim(ctx, redisKey, 0, r.maxLen)
    pipe.Expire(ctx, redisKey, r.retentionDur)
    if _, err := pipe.Exec(ctx); err != nil {
        logging.Sugar().Warnw("redis write", "err", err)
    }
    return nil
}

// ReadAll fetches all chunks from Redis newest→oldest, reverses to
// oldest→newest order, and returns deep copies.
func (r *redisStore) ReadAll() [][]byte {
    ctx := context.Background()
    vals, err := r.cli.LRange(ctx, redisKey, 0, -1).Result()
    if err != nil {
        logging.Sugar().Warnw("redis read", "err", err)
        return nil
    }
    // Reverse slice to chronological order and copy bytes.
    n := len(vals)
    out := make([][]byte, n)
    for i := 0; i < n; i++ {
        raw := []byte(vals[n-1-i])
        out[i] = append([]byte(nil), raw...)
    }
    return out
}
