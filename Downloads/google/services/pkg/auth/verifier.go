// Package auth verifies OAuth2 / IAP JWTs and exposes an interface for
// extracting the caller identity from a request.
package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is the minimal claim set FinGuard services require.
type Claims struct {
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	TenantID string   `json:"tenant_id"`
	jwt.RegisteredClaims
}

// HasRole reports whether claims contain role.
func (c *Claims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// Verifier validates a Bearer token.
type Verifier interface {
	Verify(ctx context.Context, token string) (*Claims, error)
}

// HMACVerifier is a stub verifier using a shared secret; useful for tests
// and local development. In production, use JWKSVerifier against IAP / IdP.
type HMACVerifier struct {
	Secret []byte
}

// Verify parses and validates the token using HMAC-SHA256.
func (v *HMACVerifier) Verify(_ context.Context, raw string) (*Claims, error) {
	if raw == "" {
		return nil, errors.New("auth: empty token")
	}
	raw = strings.TrimPrefix(raw, "Bearer ")
	token, err := jwt.ParseWithClaims(raw, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return v.Secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("auth: parse: %w", err)
	}
	c, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("auth: invalid claims")
	}
	return c, nil
}

// AllowAll is a no-op verifier for development / mocks. NEVER use in production.
type AllowAll struct{}

// Verify returns an anonymous claim set.
func (AllowAll) Verify(_ context.Context, _ string) (*Claims, error) {
	return &Claims{Roles: []string{"anonymous"}, RegisteredClaims: jwt.RegisteredClaims{Subject: "anonymous"}}, nil
}
