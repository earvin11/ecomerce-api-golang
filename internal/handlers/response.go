package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"ecomerce-api/internal/domain"
)

type Meta struct {
	Total     int `json:"total"`
	Page      int `json:"page"`
	PerPage   int `json:"per_page"`
	LastPage  int `json:"last_page"`
	Remaining int `json:"remaining"`
}

type Response struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
	Error   any    `json:"error"`
	Meta    *Meta  `json:"meta,omitempty"`
}

func respondJSON(w http.ResponseWriter, status int, payload Response) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondSuccess(w http.ResponseWriter, status int, message string, data any) {
	respondJSON(w, status, Response{Message: message, Data: data})
}

func respondSuccessList(w http.ResponseWriter, status int, message string, data any, meta *Meta) {
	respondJSON(w, status, Response{Message: message, Data: data, Meta: meta})
}

func respondError(w http.ResponseWriter, status int, message, detail string) {
	respondJSON(w, status, Response{Message: message, Data: nil, Error: detail})
}

func respondNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func normalizeLink(link *string) *string {
	if link == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*link)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func buildMeta(total, page, pageSize int) *Meta {
	lastPage := (total + pageSize - 1) / pageSize
	if lastPage < 1 {
		lastPage = 1
	}
	remaining := total - page*pageSize
	if remaining < 0 {
		remaining = 0
	}
	return &Meta{
		Total:     total,
		Page:      page,
		PerPage:   pageSize,
		LastPage:  lastPage,
		Remaining: remaining,
	}
}

func resolveError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		respondError(w, http.StatusNotFound, "Not found", err.Error())
	case errors.Is(err, domain.ErrInvalidData):
		respondError(w, http.StatusBadRequest, "Invalid data", err.Error())
	case errors.Is(err, domain.ErrUnauthorized):
		respondError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
	default:
		respondError(w, http.StatusInternalServerError, "Internal server error", err.Error())
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		respondError(w, http.StatusBadRequest, "Bad request", "invalid JSON body: "+err.Error())
		return false
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		respondError(w, http.StatusBadRequest, "Bad request", "invalid JSON body: unexpected trailing data")
		return false
	}
	return true
}

func parseID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "Bad request", "id must be a positive integer")
		return 0, false
	}
	return id, true
}

func parsePagination(w http.ResponseWriter, r *http.Request) (int, int, bool) {
	const (
		defaultPageSize = 10
		maxPageSize     = 100
	)

	page, pageSize := 1, defaultPageSize

	if v := r.URL.Query().Get("page"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			respondError(w, http.StatusBadRequest, "Bad request", "page must be a positive integer")
			return 0, 0, false
		}
		page = n
	}

	if v := r.URL.Query().Get("page_size"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			respondError(w, http.StatusBadRequest, "Bad request", "page_size must be a positive integer")
			return 0, 0, false
		}
		if n > maxPageSize {
			n = maxPageSize
		}
		pageSize = n
	}

	return page, pageSize, true
}
