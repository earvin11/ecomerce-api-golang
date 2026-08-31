package handlers

import (
	"net/http"

	"ecomerce-api/internal/domain"
	usecases "ecomerce-api/internal/use_cases"
)

type PurchaseHandler struct {
	useCases *usecases.PurchaseUseCases
}

func NewPurchaseHandler(useCases *usecases.PurchaseUseCases) *PurchaseHandler {
	return &PurchaseHandler{useCases: useCases}
}

func (h *PurchaseHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body domain.CreatePurchase
	if !decodeJSON(w, r, &body) {
		return
	}
	if body.ProductID <= 0 {
		respondError(w, http.StatusBadRequest, "Bad request", "product_id must be a positive integer")
		return
	}
	if body.Quantity <= 0 {
		respondError(w, http.StatusBadRequest, "Bad request", "quantity must be a positive integer")
		return
	}

	purchase, err := h.useCases.Create(r.Context(), currentUserID(r), currentRole(r), &body)
	if err != nil {
		resolveError(w, err)
		return
	}
	respondSuccess(w, http.StatusCreated, "Purchase completed successfully", purchase)
}

func (h *PurchaseHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	page, pageSize, ok := parsePagination(w, r)
	if !ok {
		return
	}

	purchases, total, err := h.useCases.GetAllByUser(r.Context(), currentUserID(r), page, pageSize)
	if err != nil {
		resolveError(w, err)
		return
	}
	respondSuccessList(w, http.StatusOK, "Purchases retrieved successfully", purchases, buildMeta(total, page, pageSize))
}
