package handlers

import (
	"net/http"
	"strings"

	"ecomerce-api/internal/domain"
	usecases "ecomerce-api/internal/use_cases"
)

type RoleHandler struct {
	useCases *usecases.RoleUseCases
}

func NewRoleHandler(useCases *usecases.RoleUseCases) *RoleHandler {
	return &RoleHandler{useCases: useCases}
}

func (h *RoleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body domain.Role
	if !decodeJSON(w, r, &body) {
		return
	}
	if !validateRoleName(w, body.Name) {
		return
	}
	body.Name = strings.ToUpper(strings.TrimSpace(body.Name))

	role, err := h.useCases.Create(r.Context(), &body)
	if err != nil {
		resolveError(w, err)
		return
	}
	respondSuccess(w, http.StatusCreated, "Role created successfully", role)
}

func (h *RoleHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	page, pageSize, ok := parsePagination(w, r)
	if !ok {
		return
	}

	roles, total, err := h.useCases.GetAll(r.Context(), page, pageSize)
	if err != nil {
		resolveError(w, err)
		return
	}
	respondSuccessList(w, http.StatusOK, "Roles retrieved successfully", roles, buildMeta(total, page, pageSize))
}

func (h *RoleHandler) GetById(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	role, err := h.useCases.GetById(r.Context(), id)
	if err != nil {
		resolveError(w, err)
		return
	}
	respondSuccess(w, http.StatusOK, "Role retrieved successfully", role)
}

func (h *RoleHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	var body domain.UpdateRole
	if !decodeJSON(w, r, &body) {
		return
	}
	if body.Name != nil && !validateRoleName(w, *body.Name) {
		return
	}
	if body.Name != nil {
		trimmed := strings.ToUpper(strings.TrimSpace(*body.Name))
		body.Name = &trimmed
	}

	role, err := h.useCases.Update(r.Context(), id, &body)
	if err != nil {
		resolveError(w, err)
		return
	}
	respondSuccess(w, http.StatusOK, "Role updated successfully", role)
}

func (h *RoleHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

func validateRoleName(w http.ResponseWriter, name string) bool {
	if strings.TrimSpace(name) == "" {
		respondError(w, http.StatusBadRequest, "Bad request", "name is required and cannot be empty")
		return false
	}
	if !domain.AllowedRoles[strings.ToUpper(strings.TrimSpace(name))] {
		respondError(w, http.StatusBadRequest, "Bad request", "name must be one of: ADMIN, CUSTOMER, EDITOR")
		return false
	}
	return true
}
