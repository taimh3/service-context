package jwtc

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTComponentLifecycleAndValidation(t *testing.T) {
	t.Run("Valid Secret and Expiry", func(t *testing.T) {
		j := NewJWT("jwt")
		if j.ID() != "jwt" {
			t.Fatalf("expected ID 'jwt', got %s", j.ID())
		}
		j.InitFlags()

		if err := j.Activate(nil); err != nil {
			t.Fatalf("unexpected Activate error: %v", err)
		}
		if err := j.Stop(); err != nil {
			t.Fatalf("unexpected Stop error: %v", err)
		}
	})

	t.Run("Secret Too Short (< 32 bytes)", func(t *testing.T) {
		j := NewJWT("jwt")
		j.secret = "too-short"
		j.expireTokenInSeconds = 3600
		if err := j.Activate(nil); err == nil {
			t.Fatal("expected error for short secret")
		}
	})

	t.Run("Expiry Too Short (<= 60s)", func(t *testing.T) {
		j := NewJWT("jwt")
		j.secret = "this-is-a-valid-32-byte-secret!!"
		j.expireTokenInSeconds = 30
		if err := j.Activate(nil); err == nil {
			t.Fatal("expected error for short expiry")
		}
	})
}

func TestJWTIssueAndParseToken(t *testing.T) {
	j := NewJWT("jwt")
	j.secret = "this-is-a-valid-32-byte-secret!!"
	j.expireTokenInSeconds = 3600

	ctx := context.Background()

	tokenStr, expSecs, err := j.IssueToken(ctx, "token-id-123", "user-456")
	if err != nil {
		t.Fatalf("IssueToken error: %v", err)
	}
	if tokenStr == "" || expSecs != 3600 {
		t.Fatalf("invalid token output: token=%s, expSecs=%d", tokenStr, expSecs)
	}

	claims, err := j.ParseToken(ctx, tokenStr)
	if err != nil {
		t.Fatalf("ParseToken error: %v", err)
	}
	if claims.ID != "token-id-123" {
		t.Fatalf("expected ID 'token-id-123', got '%s'", claims.ID)
	}
	if claims.Subject != "user-456" {
		t.Fatalf("expected Subject 'user-456', got '%s'", claims.Subject)
	}
}

func TestJWTParseInvalidToken(t *testing.T) {
	j := NewJWT("jwt")
	j.secret = "this-is-a-valid-32-byte-secret!!"
	j.expireTokenInSeconds = 3600

	ctx := context.Background()

	t.Run("Malformed token string", func(t *testing.T) {
		_, err := j.ParseToken(ctx, "invalid.token.string")
		if err == nil {
			t.Fatal("expected error parsing malformed token")
		}
	})

	t.Run("Token signed with different secret", func(t *testing.T) {
		otherJwt := NewJWT("other")
		otherJwt.secret = "another-valid-32-byte-secret-key"
		otherJwt.expireTokenInSeconds = 3600

		otherToken, _, _ := otherJwt.IssueToken(ctx, "tok1", "sub1")

		_, err := j.ParseToken(ctx, otherToken)
		if err == nil {
			t.Fatal("expected error parsing token with mismatched secret")
		}
	})

	t.Run("Expired token", func(t *testing.T) {
		claims := jwt.RegisteredClaims{
			Subject:   "user1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			ID:        "exp-id",
		}
		tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, _ := tok.SignedString([]byte(j.secret))

		_, err := j.ParseToken(ctx, signed)
		if err == nil {
			t.Fatal("expected error parsing expired token")
		}
	})
}
