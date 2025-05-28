// internal/gateway/auth.go
// Common authentication helpers for the gateway.  Supports two modes:
//  1. Static bearer token (shared secret) – very cheap check for internal
//     clusters.  Enabled when Config.AuthToken is non-empty.
//  2. JWT HMAC-SHA256 token – validates signature, issuer and expiry via
//     pkg/auth.Verifier when Config.JWTSecret is set (takes precedence over
//     plain AuthToken).
//
// The gRPC server side registers unary and stream interceptors that call
// validateBearer().  The HTTP listener attaches a standard middleware to
// protect /ws and other endpoints.
package gateway

import (
	"context"
	"net/http"
	"strings"

	"github.com/Voskan/flarego/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// JWTConfig optionally enables JWT auth.
type JWTConfig struct {
    Secret []byte // HMAC secret; if nil JWT auth is disabled
    Issuer string // expected iss claim; empty means any issuer accepted
}

// validateBearer validates Authorization header against cfg.AuthToken or JWT.
func (s *Server) validateBearer(token string) error {
    if strings.HasPrefix(token, "Bearer ") {
        token = strings.TrimPrefix(token, "Bearer ")
    }
    // Prefer JWT validation when enabled.
    if len(s.jwt.secret) > 0 {
        _, err := s.jwt.verifier.ParseAndVerify(token)
        return err
    }
    if s.cfg.AuthToken == "" {
        return nil // auth disabled
    }
    if token != s.cfg.AuthToken {
        return ErrInvalidToken
    }
    return nil
}

// installInterceptors attaches auth interceptors to the gRPC server.
func (s *Server) installInterceptors() {
    s.grpcSrv = grpc.NewServer(
        grpc.StreamInterceptor(s.streamAuthInterceptor()),
        grpc.UnaryInterceptor(s.unaryAuthInterceptor()),
    )
}

func (s *Server) unaryAuthInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        if err := s.authFromContext(ctx); err != nil {
            return nil, err
        }
        return handler(ctx, req)
    }
}

func (s *Server) streamAuthInterceptor() grpc.StreamServerInterceptor {
    return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
        if err := s.authFromContext(ss.Context()); err != nil {
            return err
        }
        return handler(srv, ss)
    }
}

func (s *Server) authFromContext(ctx context.Context) error {
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return ErrUnauthenticated
    }
    vals := md.Get("authorization")
    if len(vals) == 0 {
        return ErrUnauthenticated
    }
    return s.validateBearer(vals[0])
}

// HTTPAuthMiddleware wraps an http.Handler and enforces bearer auth.
func (s *Server) HTTPAuthMiddleware(next http.Handler) http.Handler {
    if s.cfg.AuthToken == "" && len(s.jwt.secret) == 0 {
        return next // auth disabled
    }
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if err := s.validateBearer(token); err != nil {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}

// error definitions --------------------------------------------------------
var (
    ErrUnauthenticated = status.Error(codes.Unauthenticated, "missing auth token")
    ErrInvalidToken    = status.Error(codes.PermissionDenied, "invalid auth token")
)

// internal JWT helper ------------------------------------------------------

type jwtHelper struct {
    secret   []byte
    verifier *auth.Verifier
}

func newJWTHelper(cfg JWTConfig) jwtHelper {
    if len(cfg.Secret) == 0 {
        return jwtHelper{}
    }
    return jwtHelper{
        secret:   cfg.Secret,
        verifier: auth.NewVerifier(cfg.Secret, cfg.Issuer),
    }
}
