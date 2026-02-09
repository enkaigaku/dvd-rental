package handler

import (
	"context"
	"net/http"
	"time"

	storev1 "github.com/tokyoyuan/dvd-rental/gen/proto/store/v1"
)

// StoreHandler handles store management endpoints.
type StoreHandler struct {
	storeClient storev1.StoreServiceClient
}

// NewStoreHandler creates a new StoreHandler.
func NewStoreHandler(storeClient storev1.StoreServiceClient) *StoreHandler {
	return &StoreHandler{storeClient: storeClient}
}

// --- JSON models ---

type storeResponse struct {
	StoreID        int32  `json:"store_id"`
	ManagerStaffID int32  `json:"manager_staff_id"`
	AddressID      int32  `json:"address_id"`
	LastUpdate     string `json:"last_update"`
}

type storeListResponse struct {
	Stores     []storeResponse `json:"stores"`
	TotalCount int32           `json:"total_count"`
}

type createStoreRequest struct {
	ManagerStaffID int32 `json:"manager_staff_id"`
	AddressID      int32 `json:"address_id"`
}

type updateStoreRequest struct {
	ManagerStaffID int32 `json:"manager_staff_id"`
	AddressID      int32 `json:"address_id"`
}

func storeToResponse(s *storev1.Store) storeResponse {
	return storeResponse{
		StoreID:        s.GetStoreId(),
		ManagerStaffID: s.GetManagerStaffId(),
		AddressID:      s.GetAddressId(),
		LastUpdate:     s.GetLastUpdate().AsTime().Format(time.RFC3339),
	}
}

// ListStores returns a paginated list of stores.
func (h *StoreHandler) ListStores(w http.ResponseWriter, r *http.Request) {
	pageSize, page := parsePagination(r)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.storeClient.ListStores(ctx, &storev1.ListStoresRequest{
		PageSize: pageSize,
		Page:     page,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	stores := make([]storeResponse, len(resp.GetStores()))
	for i, s := range resp.GetStores() {
		stores[i] = storeToResponse(s)
	}

	writeJSON(w, http.StatusOK, storeListResponse{
		Stores:     stores,
		TotalCount: resp.GetTotalCount(),
	})
}

// GetStore returns a single store by ID.
func (h *StoreHandler) GetStore(w http.ResponseWriter, r *http.Request) {
	storeID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid store id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	store, err := h.storeClient.GetStore(ctx, &storev1.GetStoreRequest{
		StoreId: storeID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, storeToResponse(store))
}

// CreateStore creates a new store.
func (h *StoreHandler) CreateStore(w http.ResponseWriter, r *http.Request) {
	var req createStoreRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	store, err := h.storeClient.CreateStore(ctx, &storev1.CreateStoreRequest{
		ManagerStaffId: req.ManagerStaffID,
		AddressId:      req.AddressID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, storeToResponse(store))
}

// UpdateStore updates an existing store.
func (h *StoreHandler) UpdateStore(w http.ResponseWriter, r *http.Request) {
	storeID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid store id")
		return
	}

	var req updateStoreRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	store, err := h.storeClient.UpdateStore(ctx, &storev1.UpdateStoreRequest{
		StoreId:        storeID,
		ManagerStaffId: req.ManagerStaffID,
		AddressId:      req.AddressID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, storeToResponse(store))
}

// DeleteStore deletes a store by ID.
func (h *StoreHandler) DeleteStore(w http.ResponseWriter, r *http.Request) {
	storeID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid store id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, err = h.storeClient.DeleteStore(ctx, &storev1.DeleteStoreRequest{
		StoreId: storeID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
