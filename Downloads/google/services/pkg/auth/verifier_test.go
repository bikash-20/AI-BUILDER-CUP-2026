package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestHMACVerifierRoundTrip(t *testing.T) {
	secret := []byte("dev-secret")
	v := &HMACVerifier{Secret: secret}

	claims := &Claims{
		Email: "analyst@example.com",
		Roles: []string{"analyst"},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "u-1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	got, err := v.Verify(context.Background(), "Bearer "+tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if got.Subject != "u-1" {
		t.Errorf("Subject = %q", got.Subject)
	}
	if !got.HasRole("analyst") {
		t.Error("expected role analyst")
	}
}

func TestHMACVerifierEmptyToken(t *testing.T) {
	v := &HMACVerifier{Secret: []byte("s")}
	if _, err := v.Verify(context.Background(), ""); err == nil {
		t.Fatal("expected error on empty token")
	}
}

func TestClaimsHasRole(t *testing.T) {
	c := &Claims{Roles: []string{"analyst", "reviewer"}}
	if !c.HasRole("analyst") {
		t.Error("expected analyst")
	}
	if c.HasRole("admin") {
		t.Error("did not expect admin")
	}
}

func TestAllowAllReturnsAnonymous(t *testing.T) {
	c, err := AllowAll{}.Verify(context.Background(), "")
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if c.Subject != "anonymous" {
		t.Errorf("Subject = %q", c.Subject)
	}
}
