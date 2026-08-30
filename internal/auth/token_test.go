package auth

import (
	"testing"
	"time"

	"ecomerce-api/internal/domain"
)

func TestTokenServiceRoundTrip(t *testing.T) {
	svc := NewTokenService("test-secret", 30*time.Minute, 720*time.Hour)
	claims := domain.TokenClaims{Username: "admin", Role: "ADMIN"}

	access, err := svc.SignAccess(claims)
	if err != nil {
		t.Fatalf("sign access: %v", err)
	}
	got, err := svc.ParseAccess(access)
	if err != nil {
		t.Fatalf("parse access: %v", err)
	}
	if got.Username != claims.Username || got.Role != claims.Role {
		t.Errorf("access claims: got %+v, want %+v", got, claims)
	}

	refresh, err := svc.SignRefresh(claims)
	if err != nil {
		t.Fatalf("sign refresh: %v", err)
	}
	gotRefresh, err := svc.ParseRefresh(refresh)
	if err != nil {
		t.Fatalf("parse refresh: %v", err)
	}
	if gotRefresh.Username != claims.Username || gotRefresh.Role != claims.Role {
		t.Errorf("refresh claims: got %+v, want %+v", gotRefresh, claims)
	}
}

func TestTokenServiceRejectsExpired(t *testing.T) {
	svc := NewTokenService("test-secret", -time.Second, -time.Second)
	token, err := svc.SignAccess(domain.TokenClaims{Username: "u", Role: "ADMIN"})
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if _, err := svc.ParseAccess(token); err == nil {
		t.Error("expected error for expired token")
	}
}

func TestTokenServiceRejectsTamperedToken(t *testing.T) {
	svc := NewTokenService("test-secret", 30*time.Minute, 720*time.Hour)
	token, err := svc.SignAccess(domain.TokenClaims{Username: "u", Role: "ADMIN"})
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if _, err := svc.ParseAccess(token + "x"); err == nil {
		t.Error("expected error for tampered token")
	}
}
