package handlers

import (
	"net/http"
	"regexp"
	"strings"

	"ecomerce-api/internal/domain"
	usecases "ecomerce-api/internal/use_cases"
)

var emailRegex = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

type UserHandler struct {
	useCases *usecases.UserUseCases
}

func NewUserHandler(useCases *usecases.UserUseCases) *UserHandler {
	return &UserHandler{useCases: useCases}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body domain.CreateUser
	if !decodeJSON(w, r, &body) {
		return
	}
	if !validateUserFields(w, body.Email, body.Password, body.Username, body.RoleID) {
		return
	}
	body.Email = strings.ToLower(strings.TrimSpace(body.Email))
	body.Username = strings.TrimSpace(body.Username)
	body.Photo = normalizeLink(body.Photo)

	user, err := h.useCases.Create(r.Context(), &body)
	if err != nil {
		resolveError(w, err)
		return
	}
	respondSuccess(w, http.StatusCreated, "User created successfully", user)
}

func (h *UserHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	page, pageSize, ok := parsePagination(w, r)
	if !ok {
		return
	}

	users, total, err := h.useCases.GetAll(r.Context(), page, pageSize)
	if err != nil {
		resolveError(w, err)
		return
	}
	respondSuccessList(w, http.StatusOK, "Users retrieved successfully", users, buildMeta(total, page, pageSize))
}

func (h *UserHandler) GetById(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	user, err := h.useCases.GetById(r.Context(), id)
	if err != nil {
		resolveError(w, err)
		return
	}
	respondSuccess(w, http.StatusOK, "User retrieved successfully", user)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	var body domain.UpdateUser
	if !decodeJSON(w, r, &body) {
		return
	}
	if body.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*body.Email))
		if !emailRegex.MatchString(email) {
			respondError(w, http.StatusBadRequest, "Bad request", "email must be a valid email address")
			return
		}
		body.Email = &email
	}
	if body.Password != nil && len(*body.Password) < 8 {
		respondError(w, http.StatusBadRequest, "Bad request", "password must be at least 8 characters")
		return
	}
	if body.Username != nil {
		username := strings.TrimSpace(*body.Username)
		if len(username) < 3 {
			respondError(w, http.StatusBadRequest, "Bad request", "username must be at least 3 characters")
			return
		}
		body.Username = &username
	}
	if body.RoleID != nil && *body.RoleID <= 0 {
		respondError(w, http.StatusBadRequest, "Bad request", "role_id must be a positive integer")
		return
	}
	if body.Photo.Set {
		body.Photo.Value = normalizeLink(body.Photo.Value)
	}

	user, err := h.useCases.Update(r.Context(), id, &body)
	if err != nil {
		resolveError(w, err)
		return
	}
	respondSuccess(w, http.StatusOK, "User updated successfully", user)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

func validateUserFields(w http.ResponseWriter, email, password, username string, roleID int) bool {
	switch {
	case !emailRegex.MatchString(email):
		respondError(w, http.StatusBadRequest, "Bad request", "email must be a valid email address")
		return false
	case len(password) < 8:
		respondError(w, http.StatusBadRequest, "Bad request", "password must be at least 8 characters")
		return false
	case len(strings.TrimSpace(username)) < 3:
		respondError(w, http.StatusBadRequest, "Bad request", "username must be at least 3 characters")
		return false
	case roleID <= 0:
		respondError(w, http.StatusBadRequest, "Bad request", "role_id must be a positive integer")
		return false
	}
	return true
}
