// internal/gateway/server.go
// Package gateway exposes a gRPC front‑door for agents and a fan‑out hub for
// UI subscribers (WebSocket, gRPC‑web, etc.).  The server is intentionally
// lightweight; retention and alerting are delegated to pluggable components in
// sibling packages.
package gateway

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
	"time"

	"github.com/Voskan/flarego/internal/gateway/retention"
	"github.com/Voskan/flarego/internal/logging"
	agentpb "github.com/Voskan/flarego/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Config parameterises a Gateway Server.
type Config struct {
    ListenAddr   string        // host:port to bind
    TLSConfig    *tls.Config   // nil to serve over plaintext
    AuthToken    string        // optional static bearer token ("" means open)
    RetentionDur time.Duration // how long to keep a chunk in memory (0 => 15m)
    MaxClients   int           // soft cap for connected subscribers
    TLSCertPath  string        // path to TLS certificate (PEM)
    TLSKeyPath   string        // path to TLS key (PEM)
}

// Server implements the generated gRPC service and fans‑out chunks to all
// attached UI subscribers (via Subscribe()) while writing them to a Retention
// Store for replay.
type Server struct {
    agentpb.UnimplementedGatewayServiceServer
    agentpb.UnimplementedUIServiceServer

    cfg     Config
    store   retention.Store
    subsMu  sync.RWMutex
    subs    map[chan []byte]struct{}
    grpcSrv *grpc.Server
    jwt     jwtHelper
}

// New returns a ready‑to‑serve Gateway.  The caller must invoke ListenAndServe.
func New(cfg Config) (*Server, error) {
    if cfg.RetentionDur == 0 {
        cfg.RetentionDur = 15 * time.Minute
    }
    s := &Server{
        cfg:   cfg,
        store: retention.NewInMem(cfg.RetentionDur),
        subs:  make(map[chan []byte]struct{}),
    }

    var opts []grpc.ServerOption
    if cfg.TLSConfig != nil {
        opts = append(opts, grpc.Creds(credentials.NewTLS(cfg.TLSConfig)))
    }

    // Add CORS interceptor
    opts = append(opts, grpc.ChainUnaryInterceptor(
        func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
            md, ok := metadata.FromIncomingContext(ctx)
            if ok {
                md.Set("Access-Control-Allow-Origin", "*")
                md.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
                md.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
                ctx = metadata.NewOutgoingContext(ctx, md)
            }
            return handler(ctx, req)
        },
    ))
    opts = append(opts, grpc.ChainStreamInterceptor(
        func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
            ctx := ss.Context()
            md, ok := metadata.FromIncomingContext(ctx)
            if ok {
                md.Set("Access-Control-Allow-Origin", "*")
                md.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
                md.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
                ctx = metadata.NewOutgoingContext(ctx, md)
            }
            return handler(srv, &wrappedServerStream{ServerStream: ss, ctx: ctx})
        },
    ))

    s.grpcSrv = grpc.NewServer(opts...)
    agentpb.RegisterGatewayServiceServer(s.grpcSrv, s)
    agentpb.RegisterUIServiceServer(s.grpcSrv, s)
    return s, nil
}

// wrappedServerStream wraps grpc.ServerStream to override Context()
type wrappedServerStream struct {
    grpc.ServerStream
    ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
    return w.ctx
}

// ListenAndServe blocks, serving the gRPC API until ctx is cancelled.
func (s *Server) ListenAndServe(ctx context.Context) error {
    ln, err := net.Listen("tcp", s.cfg.ListenAddr)
    if err != nil {
        return err
    }

    go func() {
        <-ctx.Done()
        // GracefulStop drains existing RPCs; Close closes listener.
        s.grpcSrv.GracefulStop()
        _ = ln.Close()
    }()

    logging.Sugar().Infow("gateway listening", "addr", ln.Addr().String())
    return s.grpcSrv.Serve(ln)
}

// Stream is the hot path: agents push FlamegraphChunk frames continuously.
func (s *Server) Stream(stream agentpb.GatewayService_StreamServer) error {
    // Optional bearer‑token auth.
    if s.cfg.AuthToken != "" {
        md, ok := metadata.FromIncomingContext(stream.Context())
        if !ok || len(md.Get("authorization")) == 0 {
            return status.Error(codes.Unauthenticated, "missing auth token")
        }
        tok := md.Get("authorization")[0]
        expected := "Bearer " + s.cfg.AuthToken
        if tok != expected {
            return status.Error(codes.PermissionDenied, "invalid auth token")
        }
    }

    // Read chunks until EOF.
    for {
        chunk, err := stream.Recv()
        if err != nil {
            if status.Code(err) == codes.Canceled || status.Code(err) == codes.Unavailable {
                return nil // client disconnected
            }
            logging.Sugar().Warnw("stream recv", "err", err)
            return err
        }
        s.handleChunk(chunk.Payload)
    }
}

// StreamFlamegraphs is the UI service endpoint that streams flamegraph chunks to clients.
func (s *Server) StreamFlamegraphs(req *emptypb.Empty, stream agentpb.UIService_StreamFlamegraphsServer) error {
    // Optional bearer‑token auth.
    if s.cfg.AuthToken != "" {
        md, ok := metadata.FromIncomingContext(stream.Context())
        if !ok || len(md.Get("authorization")) == 0 {
            return status.Error(codes.Unauthenticated, "missing auth token")
        }
        tok := md.Get("authorization")[0]
        expected := "Bearer " + s.cfg.AuthToken
        if tok != expected {
            return status.Error(codes.PermissionDenied, "invalid auth token")
        }
    }

    // Create a channel for this subscriber.
    ch := make(chan []byte, 100) // buffered to avoid blocking the gateway
    s.subsMu.Lock()
    s.subs[ch] = struct{}{}
    s.subsMu.Unlock()

    // Clean up when the client disconnects.
    defer func() {
        s.subsMu.Lock()
        delete(s.subs, ch)
        s.subsMu.Unlock()
        close(ch)
    }()

    // Send initial data from retention store.
    for _, data := range s.store.ReadAll() {
        if err := stream.Send(&agentpb.FlamegraphChunk{Payload: data}); err != nil {
            return err
        }
    }

    // Stream new chunks until client disconnects.
    for data := range ch {
        if err := stream.Send(&agentpb.FlamegraphChunk{Payload: data}); err != nil {
            return err
        }
    }

    return nil
}

// handleChunk writes to store and broadcasts to subscribers.
func (s *Server) handleChunk(data []byte) {
    // Persist in ring buffer.
    if err := s.store.Write(data); err != nil {
        logging.Sugar().Warnw("retention write", "err", err)
    }

    // Non‑blocking fan‑out.
    s.subsMu.RLock()
    for ch := range s.subs {
        select {
        case ch <- data:
        default:
            // Skip slow consumer to avoid head‑of‑line blocking.
            logging.Sugar().Debug("dropping chunk to slow subscriber")
        }
    }
    s.subsMu.RUnlock()
}

// Subscribe registers a UI client.  The caller must drain the returned channel
// and invoke the unregister func when done.
func (s *Server) Subscribe() (ch chan []byte, unregister func()) {
    ch = make(chan []byte, 100) // buffered to avoid blocking the gateway
    s.subsMu.Lock()
    s.subs[ch] = struct{}{}
    s.subsMu.Unlock()

    unregister = func() {
        s.subsMu.Lock()
        delete(s.subs, ch)
        s.subsMu.Unlock()
        close(ch)
    }

    return ch, unregister
}

// Logger returns the *zap.Logger used by the server (delegates to global).
func (s *Server) Logger() *zap.Logger { return logging.Logger() }
