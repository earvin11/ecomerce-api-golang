package handlers

import (
	"context"
	"net/http"
)

type HealthChecker interface {
	Ping(ctx context.Context) error
}

type Handler struct {
	category      *CategoryHandler
	product       *ProductHandler
	role          *RoleHandler
	user          *UserHandler
	auth          *AuthHandler
	middleware    *AuthMiddleware
	healthChecker HealthChecker
}

func NewHandler(category *CategoryHandler, product *ProductHandler, role *RoleHandler, user *UserHandler, auth *AuthHandler, middleware *AuthMiddleware, healthChecker HealthChecker) *Handler {
	return &Handler{
		category:      category,
		product:       product,
		role:          role,
		user:          user,
		auth:          auth,
		middleware:    middleware,
		healthChecker: healthChecker,
	}
}

func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("POST /api/v1/auth/login", h.auth.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", h.auth.Refresh)

	mux.Handle("GET /api/v1/categories", h.middleware.Require(h.category.GetAll))
	mux.Handle("POST /api/v1/categories", h.middleware.Require(h.category.Create))
	mux.Handle("GET /api/v1/categories/{id}", h.middleware.Require(h.category.GetById))
	mux.Handle("PATCH /api/v1/categories/{id}", h.middleware.Require(h.category.Update))
	mux.Handle("DELETE /api/v1/categories/{id}", h.middleware.Require(h.category.Delete))

	mux.Handle("GET /api/v1/products", h.middleware.Require(h.product.GetAll))
	mux.Handle("POST /api/v1/products", h.middleware.Require(h.product.Create))
	mux.Handle("GET /api/v1/products/{id}", h.middleware.Require(h.product.GetById))
	mux.Handle("PATCH /api/v1/products/{id}", h.middleware.Require(h.product.Update))
	mux.Handle("DELETE /api/v1/products/{id}", h.middleware.Require(h.product.Delete))

	mux.Handle("GET /api/v1/roles", h.middleware.Require(h.role.GetAll))
	mux.Handle("POST /api/v1/roles", h.middleware.Require(h.role.Create))
	mux.Handle("GET /api/v1/roles/{id}", h.middleware.Require(h.role.GetById))
	mux.Handle("PATCH /api/v1/roles/{id}", h.middleware.Require(h.role.Update))
	mux.Handle("DELETE /api/v1/roles/{id}", h.middleware.Require(h.role.Delete))

	mux.Handle("GET /api/v1/users", h.middleware.Require(h.user.GetAll))
	mux.Handle("POST /api/v1/users", h.middleware.Require(h.user.Create))
	mux.Handle("GET /api/v1/users/{id}", h.middleware.Require(h.user.GetById))
	mux.Handle("PATCH /api/v1/users/{id}", h.middleware.Require(h.user.Update))
	mux.Handle("DELETE /api/v1/users/{id}", h.middleware.Require(h.user.Delete))

	mux.HandleFunc("/", h.notFound)

	return mux
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	if err := h.healthChecker.Ping(r.Context()); err != nil {
		respondError(w, http.StatusServiceUnavailable, "Service unavailable", "database connection failed")
		return
	}
	respondSuccess(w, http.StatusOK, "Service is healthy", map[string]string{"status": "ok"})
}

func (h *Handler) notFound(w http.ResponseWriter, r *http.Request) {
	respondError(w, http.StatusNotFound, "Not found", "the requested resource does not exist")
}
