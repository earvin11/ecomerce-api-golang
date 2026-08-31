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
	ctxUserID   contextKey = "user_id"
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
		ctx = context.WithValue(ctx, ctxUserID, claims.UserID)
		next(w, r.WithContext(ctx))
	}
}

func (m *AuthMiddleware) RequireRoles(roles ...string) func(http.HandlerFunc) http.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if _, ok := allowed[currentRole(r)]; !ok {
				respondError(w, http.StatusForbidden, "Forbidden", "insufficient permissions")
				return
			}
			next(w, r)
		}
	}
}

func currentUserID(r *http.Request) int {
	id, _ := r.Context().Value(ctxUserID).(int)
	return id
}

func currentRole(r *http.Request) string {
	role, _ := r.Context().Value(ctxRole).(string)
	return role
}
