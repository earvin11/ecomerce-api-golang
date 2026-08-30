package handlers

import (
	"net/http"
	"strings"

	"ecomerce-api/internal/domain"
	usecases "ecomerce-api/internal/use_cases"
)

type AuthHandler struct {
	useCases *usecases.AuthUseCases
}

func NewAuthHandler(useCases *usecases.AuthUseCases) *AuthHandler {
	return &AuthHandler{useCases: useCases}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body domain.Credentials
	if !decodeJSON(w, r, &body) {
		return
	}
	body.Username = strings.TrimSpace(body.Username)
	if body.Username == "" || body.Password == "" {
		respondError(w, http.StatusBadRequest, "Bad request", "username and password are required")
		return
	}

	pair, err := h.useCases.Login(r.Context(), &body)
	if err != nil {
		resolveError(w, err)
		return
	}
	respondSuccess(w, http.StatusOK, "Login successful", pair)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var body domain.RefreshRequest
	if !decodeJSON(w, r, &body) {
		return
	}
	if strings.TrimSpace(body.RefreshToken) == "" {
		respondError(w, http.StatusBadRequest, "Bad request", "refresh_token is required")
		return
	}

	pair, err := h.useCases.Refresh(r.Context(), body.RefreshToken)
	if err != nil {
		resolveError(w, err)
		return
	}
	respondSuccess(w, http.StatusOK, "Session refreshed successfully", pair)
}
