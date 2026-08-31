package handlers

import (
	"net/http"
	"strings"

	"ecomerce-api/internal/domain"
	usecases "ecomerce-api/internal/use_cases"
)

type ProductHandler struct {
	useCases *usecases.ProductUseCases
}

func NewProductHandler(useCases *usecases.ProductUseCases) *ProductHandler {
	return &ProductHandler{useCases: useCases}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body domain.Product
	if !decodeJSON(w, r, &body) {
		return
	}
	if !validateProduct(w, &body) {
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	body.Img = normalizeLink(body.Img)

	product, err := h.useCases.Create(r.Context(), &body, currentUserID(r))
	if err != nil {
		resolveError(w, err)
		return
	}
	respondSuccess(w, http.StatusCreated, "Product created successfully", product)
}

func (h *ProductHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	page, pageSize, ok := parsePagination(w, r)
	if !ok {
		return
	}

	products, total, err := h.useCases.GetAll(r.Context(), page, pageSize)
	if err != nil {
		resolveError(w, err)
		return
	}
	respondSuccessList(w, http.StatusOK, "Products retrieved successfully", products, buildMeta(total, page, pageSize))
}

func (h *ProductHandler) GetById(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	product, err := h.useCases.GetById(r.Context(), id)
	if err != nil {
		resolveError(w, err)
		return
	}
	respondSuccess(w, http.StatusOK, "Product retrieved successfully", product)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	var body domain.UpdateProduct
	if !decodeJSON(w, r, &body) {
		return
	}
	if body.Name != nil && strings.TrimSpace(*body.Name) == "" {
		respondError(w, http.StatusBadRequest, "Bad request", "name cannot be empty")
		return
	}
	if body.Name != nil {
		trimmed := strings.TrimSpace(*body.Name)
		body.Name = &trimmed
	}
	if body.Price != nil && *body.Price < 0 {
		respondError(w, http.StatusBadRequest, "Bad request", "price cannot be negative")
		return
	}
	if body.Category != nil && *body.Category <= 0 {
		respondError(w, http.StatusBadRequest, "Bad request", "category must be a positive integer")
		return
	}
	if body.Quantity != nil && *body.Quantity < 0 {
		respondError(w, http.StatusBadRequest, "Bad request", "quantity cannot be negative")
		return
	}
	if body.Img.Set {
		body.Img.Value = normalizeLink(body.Img.Value)
	}

	product, err := h.useCases.Update(r.Context(), id, &body, currentUserID(r))
	if err != nil {
		resolveError(w, err)
		return
	}
	respondSuccess(w, http.StatusOK, "Product updated successfully", product)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	if err := h.useCases.Delete(r.Context(), id); err != nil {
		resolveError(w, err)
		return
	}
	respondNoContent(w)
}

func validateProduct(w http.ResponseWriter, p *domain.Product) bool {
	switch {
	case strings.TrimSpace(p.Name) == "":
		respondError(w, http.StatusBadRequest, "Bad request", "name is required and cannot be empty")
		return false
	case p.Price < 0:
		respondError(w, http.StatusBadRequest, "Bad request", "price cannot be negative")
		return false
	case p.Category <= 0:
		respondError(w, http.StatusBadRequest, "Bad request", "category must be a positive integer")
		return false
	case p.Quantity < 0:
		respondError(w, http.StatusBadRequest, "Bad request", "quantity cannot be negative")
		return false
	}
	return true
}
