// pkg/auth/jwt.go
// Lightweight HMAC‑SHA256 JWT signer / verifier used by both agent and gateway
// for authentication.  The implementation deliberately avoids advanced JWT
// conventions (kid, JWKs) to keep the dependency surface minimal.
//
// External dependency: github.com/golang-jwt/jwt/v5 (MIT).
package auth

import (
	"errors"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

// Signer produces short‑lived tokens for agents.
type Signer struct {
    secret     []byte
    issuer     string
    ttl        time.Duration
    clock      func() time.Time // injection point for tests
}

// NewSigner returns a Signer with given secret, issuer claim and TTL.
func NewSigner(secret []byte, issuer string, ttl time.Duration) *Signer {
    if ttl <= 0 {
        ttl = 15 * time.Minute
    }
    return &Signer{secret: secret, issuer: issuer, ttl: ttl, clock: time.Now}
}

// Claims returns standard claims for a new token.
func (s *Signer) Claims(subject string, extra map[string]any) jwt.MapClaims {
    now := s.clock()
    claims := jwt.MapClaims{
        "iss": s.issuer,
        "sub": subject,
        "iat": now.Unix(),
        "exp": now.Add(s.ttl).Unix(),
    }
    for k, v := range extra {
        claims[k] = v
    }
    return claims
}

// Sign produces a JWT string.
func (s *Signer) Sign(claims jwt.MapClaims) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(s.secret)
}

// Verifier validates HMAC‑signed tokens.
type Verifier struct {
    secret []byte
    issuer string
    clock  func() time.Time
}

// NewVerifier constructs a verifier with expected issuer.
func NewVerifier(secret []byte, issuer string) *Verifier {
    return &Verifier{secret: secret, issuer: issuer, clock: time.Now}
}

var (
    ErrInvalidToken  = errors.New("invalid token")
    ErrExpiredToken  = errors.New("token expired")
    ErrIssuerMismatch = errors.New("issuer mismatch")
)

// ParseAndVerify parses tokenStr and returns claims after validating signature,
// expiry and issuer.
func (v *Verifier) ParseAndVerify(tokenStr string) (jwt.MapClaims, error) {
    token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
        if t.Method != jwt.SigningMethodHS256 {
            return nil, ErrInvalidToken
        }
        return v.secret, nil
    }, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
    if err != nil {
        if errors.Is(err, jwt.ErrTokenExpired) {
            return nil, ErrExpiredToken
        }
        return nil, ErrInvalidToken
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok || !token.Valid {
        return nil, ErrInvalidToken
    }
    if v.issuer != "" && claims["iss"] != v.issuer {
        return nil, ErrIssuerMismatch
    }
    return claims, nil
}
