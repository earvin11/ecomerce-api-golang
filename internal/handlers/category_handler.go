package handlers

import (
	"net/http"
	"strings"

	"ecomerce-api/internal/domain"
	usecases "ecomerce-api/internal/use_cases"
)

type CategoryHandler struct {
	useCases *usecases.CategoryUseCases
}

func NewCategoryHandler(useCases *usecases.CategoryUseCases) *CategoryHandler {
	return &CategoryHandler{useCases: useCases}
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body domain.Category
	if !decodeJSON(w, r, &body) {
		return
	}
	if strings.TrimSpace(body.Name) == "" {
		respondError(w, http.StatusBadRequest, "Bad request", "name is required and cannot be empty")
		return
	}
	body.Name = strings.TrimSpace(body.Name)

	category, err := h.useCases.Create(r.Context(), &body, currentUserID(r))
	if err != nil {
		resolveError(w, err)
		return
	}
	respondSuccess(w, http.StatusCreated, "Category created successfully", category)
}

func (h *CategoryHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	page, pageSize, ok := parsePagination(w, r)
	if !ok {
		return
	}

	categories, total, err := h.useCases.GetAll(r.Context(), page, pageSize)
	if err != nil {
		resolveError(w, err)
		return
	}
	respondSuccessList(w, http.StatusOK, "Categories retrieved successfully", categories, buildMeta(total, page, pageSize))
}

func (h *CategoryHandler) GetById(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	category, err := h.useCases.GetById(r.Context(), id)
	if err != nil {
		resolveError(w, err)
		return
	}
	respondSuccess(w, http.StatusOK, "Category retrieved successfully", category)
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	var body domain.UpdateCategory
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

	category, err := h.useCases.Update(r.Context(), id, &body, currentUserID(r))
	if err != nil {
		resolveError(w, err)
		return
	}
	respondSuccess(w, http.StatusOK, "Category updated successfully", category)
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
