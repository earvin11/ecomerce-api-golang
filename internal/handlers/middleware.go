package handlers

import (
	"context"
	"net/http"
	"strings"

	authpkg "ecomerce-api/internal/auth"
)

type contextKey string

const (
	ctxUsername contextKey = "username"
	ctxRole     contextKey = "role"
)

type AuthMiddleware struct {
	tokens *authpkg.TokenService
}

func NewAuthMiddleware(tokens *authpkg.TokenService) *AuthMiddleware {
	return &AuthMiddleware{tokens: tokens}
}

func (m *AuthMiddleware) Require(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := strings.TrimSpace(r.Header.Get("Authorization"))
		if !strings.HasPrefix(header, "Bearer ") {
			respondError(w, http.StatusUnauthorized, "Unauthorized", "missing or malformed Authorization header")
			return
		}
		claims, err := m.tokens.ParseAccess(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			respondError(w, http.StatusUnauthorized, "Unauthorized", "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), ctxUsername, claims.Username)
		ctx = context.WithValue(ctx, ctxRole, claims.Role)
		next(w, r.WithContext(ctx))
	}
}
