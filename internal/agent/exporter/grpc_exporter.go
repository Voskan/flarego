// internal/agent/exporter/grpc_exporter.go
// Package exporter implements Exporter adapters used by the in‑process agent
// to transmit aggregated flame graphs to a FlareGo gateway.  The gRPC exporter
// maintains a persistent bidirectional stream and performs automatic
// reconnect with jittered exponential back‑off.
package exporter

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/cenkalti/backoff/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	agentpb "github.com/Voskan/flarego/internal/proto"
	"github.com/Voskan/flarego/pkg/flamegraph"
)

// Config defines connection parameters for the gRPC exporter.
//
// ● Addr is of the form host:port (TLS implied).  If Opts overrides creds the
//   handshake follows Opts.
// ● AuthToken, if non‑empty, is sent via gRPC metadata key "authorization"
//   as "Bearer <token>".
// ● StreamRetry controls reconnection policy; if nil a sensible default
//   (max 1 minute, factor 2, jitter) is used.
// ● FlushTimeout bounds time spent per Export call.
type Config struct {
    Addr         string
    AuthToken    string
    Opts         []grpc.DialOption
    StreamRetry  backoff.BackOff
    FlushTimeout time.Duration
}

// grpcExporter implements agent.Exporter.
type grpcExporter struct {
    cfg    Config
    client agentpb.GatewayServiceClient
    conn   *grpc.ClientConn
    stream agentpb.GatewayService_StreamClient

    closing chan struct{}
}

// NewGRPCExporter creates and connects an exporter. The call blocks until the
// first successful handshake.
func NewGRPCExporter(ctx context.Context, cfg Config) (*grpcExporter, error) {
    g := &grpcExporter{
        cfg:     cfg,
        closing: make(chan struct{}),
    }
    if cfg.StreamRetry == nil {
        bo := backoff.NewExponentialBackOff()
        bo.InitialInterval = 500 * time.Millisecond
        bo.MaxInterval = 15 * time.Second
        bo.MaxElapsedTime = time.Minute
        cfg.StreamRetry = bo
        g.cfg.StreamRetry = bo
    }
    if err := g.connect(ctx); err != nil {
        return nil, err
    }
    return g, nil
}

// Export sends one flame‑graph snapshot over the stream, retrying the stream
// if necessary.  It satisfies agent.Exporter.
func (g *grpcExporter) Export(ctx context.Context, root *flamegraph.Frame) error {
    if root == nil {
        return nil
    }
    // Marshal graph to JSON; the UI and gateway accept raw JSON blob to keep
    // proto schema stable.
    data, err := root.ToJSON()
    if err != nil {
        return err
    }

    // Ensure a stream is alive.
    if g.stream == nil {
        if err := g.connect(ctx); err != nil {
            return err
        }
    }

    // Respect flush timeout.
    to := g.cfg.FlushTimeout
    if to == 0 {
        to = 5 * time.Second
    }
    ctx, cancel := context.WithTimeout(ctx, to)
    defer cancel()

    if err := g.stream.Send(&agentpb.FlamegraphChunk{Payload: data}); err != nil {
        // Attempt reconnection once; caller may re‑invoke.
        _ = g.reconnect(ctx)
        return err
    }
    return nil
}

// Close terminates the stream and the underlying connection.
func (g *grpcExporter) Close() error {
    close(g.closing)
    if g.stream != nil {
        _ = g.stream.CloseSend()
    }
    if g.conn != nil {
        return g.conn.Close()
    }
    return nil
}

// connect dials the gateway and opens a new stream.
func (g *grpcExporter) connect(ctx context.Context) error {
    dialOpts := append([]grpc.DialOption{}, g.cfg.Opts...)
    // Default TLS unless disabled by Opts.
    hasCreds := false
    for _, o := range dialOpts {
        if _, ok := o.(grpc.CredsCallOption); ok {
            hasCreds = true
            break
        }
    }
    if !hasCreds {
        dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})))
    }
    // Block until ready.
    dialOpts = append(dialOpts, grpc.WithBlock())

    conn, err := grpc.DialContext(ctx, g.cfg.Addr, dialOpts...)
    if err != nil {
        return err
    }
    client := agentpb.NewGatewayServiceClient(conn)

    md := metadata.New(nil)
    if g.cfg.AuthToken != "" {
        md.Set("authorization", "Bearer "+g.cfg.AuthToken)
    }
    stream, err := client.Stream(metadata.NewOutgoingContext(ctx, md))
    if err != nil {
        _ = conn.Close()
        return err
    }

    g.conn = conn
    g.client = client
    g.stream = stream
    return nil
}

// reconnect closes existing resources and retries connect() respecting the
// configured back‑off policy.
func (g *grpcExporter) reconnect(ctx context.Context) error {
    if g.stream != nil {
        _ = g.stream.CloseSend()
        g.stream = nil
    }
    if g.conn != nil {
        _ = g.conn.Close()
        g.conn = nil
    }

    bo := g.cfg.StreamRetry
    bo.Reset()
    for {
        next := bo.NextBackOff()
        if next == backoff.Stop {
            return context.DeadlineExceeded
        }
        select {
        case <-time.After(next):
        case <-g.closing:
            return context.Canceled
        case <-ctx.Done():
            return ctx.Err()
        }
        if err := g.connect(ctx); err == nil {
            return nil
        }
    }
}
