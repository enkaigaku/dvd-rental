package handler

import (
	"context"
	"net/http"
	"time"

	rentalv1 "github.com/enkaigaku/dvd-rental/gen/proto/rental/v1"
)

// InventoryHandler handles inventory management endpoints.
type InventoryHandler struct {
	inventoryClient rentalv1.InventoryServiceClient
}

// NewInventoryHandler creates a new InventoryHandler.
func NewInventoryHandler(inventoryClient rentalv1.InventoryServiceClient) *InventoryHandler {
	return &InventoryHandler{inventoryClient: inventoryClient}
}

// --- JSON models ---

type inventoryResponse struct {
	InventoryID int32  `json:"inventory_id"`
	FilmID      int32  `json:"film_id"`
	StoreID     int32  `json:"store_id"`
	LastUpdate  string `json:"last_update"`
}

type inventoryListResponse struct {
	Inventory  []inventoryResponse `json:"inventory"`
	TotalCount int32               `json:"total_count"`
}

type createInventoryRequest struct {
	FilmID  int32 `json:"film_id"`
	StoreID int32 `json:"store_id"`
}

type availabilityResponse struct {
	InventoryID int32 `json:"inventory_id"`
	Available   bool  `json:"available"`
}

func inventoryToResponse(i *rentalv1.Inventory) inventoryResponse {
	return inventoryResponse{
		InventoryID: i.GetInventoryId(),
		FilmID:      i.GetFilmId(),
		StoreID:     i.GetStoreId(),
		LastUpdate:  i.GetLastUpdate().AsTime().Format(time.RFC3339),
	}
}

// ListInventory returns a paginated list of inventory items.
func (h *InventoryHandler) ListInventory(w http.ResponseWriter, r *http.Request) {
	pageSize, page := parsePagination(r)
	filmID := parseQueryInt32(r, "film_id")
	storeID := parseQueryInt32(r, "store_id")

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var resp *rentalv1.ListInventoryResponse
	var err error

	switch {
	case filmID > 0 && storeID > 0:
		resp, err = h.inventoryClient.ListAvailableInventory(ctx, &rentalv1.ListAvailableInventoryRequest{
			FilmId:   filmID,
			StoreId:  storeID,
			PageSize: pageSize,
			Page:     page,
		})
	case filmID > 0:
		resp, err = h.inventoryClient.ListInventoryByFilm(ctx, &rentalv1.ListInventoryByFilmRequest{
			FilmId:   filmID,
			PageSize: pageSize,
			Page:     page,
		})
	case storeID > 0:
		resp, err = h.inventoryClient.ListInventoryByStore(ctx, &rentalv1.ListInventoryByStoreRequest{
			StoreId:  storeID,
			PageSize: pageSize,
			Page:     page,
		})
	default:
		resp, err = h.inventoryClient.ListInventory(ctx, &rentalv1.ListInventoryRequest{
			PageSize: pageSize,
			Page:     page,
		})
	}
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	inventory := make([]inventoryResponse, len(resp.GetItems()))
	for i, inv := range resp.GetItems() {
		inventory[i] = inventoryToResponse(inv)
	}

	writeJSON(w, http.StatusOK, inventoryListResponse{
		Inventory:  inventory,
		TotalCount: resp.GetTotalCount(),
	})
}

// GetInventory returns a single inventory item by ID.
func (h *InventoryHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	inventoryID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid inventory id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	inv, err := h.inventoryClient.GetInventory(ctx, &rentalv1.GetInventoryRequest{
		InventoryId: inventoryID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, inventoryToResponse(inv))
}

// CreateInventory creates a new inventory item.
func (h *InventoryHandler) CreateInventory(w http.ResponseWriter, r *http.Request) {
	var req createInventoryRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	inv, err := h.inventoryClient.CreateInventory(ctx, &rentalv1.CreateInventoryRequest{
		FilmId:  req.FilmID,
		StoreId: req.StoreID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, inventoryToResponse(inv))
}

// DeleteInventory deletes an inventory item by ID.
func (h *InventoryHandler) DeleteInventory(w http.ResponseWriter, r *http.Request) {
	inventoryID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid inventory id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, err = h.inventoryClient.DeleteInventory(ctx, &rentalv1.DeleteInventoryRequest{
		InventoryId: inventoryID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CheckAvailability checks if an inventory item is available for rent.
func (h *InventoryHandler) CheckAvailability(w http.ResponseWriter, r *http.Request) {
	inventoryID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid inventory id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.inventoryClient.CheckInventoryAvailability(ctx, &rentalv1.CheckInventoryAvailabilityRequest{
		InventoryId: inventoryID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, availabilityResponse{
		InventoryID: inventoryID,
		Available:   resp.GetAvailable(),
	})
}
