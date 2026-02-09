package handler

import (
	"context"
	"net/http"
	"time"

	rentalv1 "github.com/enkaigaku/dvd-rental/gen/proto/rental/v1"
)

// RentalHandler handles rental management endpoints.
type RentalHandler struct {
	rentalClient rentalv1.RentalServiceClient
}

// NewRentalHandler creates a new RentalHandler.
func NewRentalHandler(rentalClient rentalv1.RentalServiceClient) *RentalHandler {
	return &RentalHandler{rentalClient: rentalClient}
}

// --- JSON models ---

type rentalResponse struct {
	RentalID    int32  `json:"rental_id"`
	RentalDate  string `json:"rental_date"`
	InventoryID int32  `json:"inventory_id"`
	CustomerID  int32  `json:"customer_id"`
	ReturnDate  string `json:"return_date,omitempty"`
	StaffID     int32  `json:"staff_id"`
	LastUpdate  string `json:"last_update"`
}

type rentalDetailResponse struct {
	rentalResponse
	FilmTitle    string `json:"film_title"`
	CustomerName string `json:"customer_name"`
	StoreID      int32  `json:"store_id"`
}

type rentalListResponse struct {
	Rentals    []rentalResponse `json:"rentals"`
	TotalCount int32            `json:"total_count"`
}

type createRentalRequest struct {
	InventoryID int32 `json:"inventory_id"`
	CustomerID  int32 `json:"customer_id"`
	StaffID     int32 `json:"staff_id"`
}

func rentalToResponse(r *rentalv1.Rental) rentalResponse {
	resp := rentalResponse{
		RentalID:    r.GetRentalId(),
		RentalDate:  r.GetRentalDate().AsTime().Format(time.RFC3339),
		InventoryID: r.GetInventoryId(),
		CustomerID:  r.GetCustomerId(),
		StaffID:     r.GetStaffId(),
		LastUpdate:  r.GetLastUpdate().AsTime().Format(time.RFC3339),
	}
	if r.GetReturnDate() != nil {
		resp.ReturnDate = r.GetReturnDate().AsTime().Format(time.RFC3339)
	}
	return resp
}

func rentalDetailToResponse(d *rentalv1.RentalDetail) rentalDetailResponse {
	return rentalDetailResponse{
		rentalResponse: rentalToResponse(d.GetRental()),
		FilmTitle:      d.GetFilmTitle(),
		CustomerName:   d.GetCustomerName(),
		StoreID:        d.GetStoreId(),
	}
}

// ListRentals returns a paginated list of rentals.
func (h *RentalHandler) ListRentals(w http.ResponseWriter, r *http.Request) {
	pageSize, page := parsePagination(r)
	customerID := parseQueryInt32(r, "customer_id")
	inventoryID := parseQueryInt32(r, "inventory_id")

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var resp *rentalv1.ListRentalsResponse
	var err error

	switch {
	case customerID > 0:
		resp, err = h.rentalClient.ListRentalsByCustomer(ctx, &rentalv1.ListRentalsByCustomerRequest{
			CustomerId: customerID,
			PageSize:   pageSize,
			Page:       page,
		})
	case inventoryID > 0:
		resp, err = h.rentalClient.ListRentalsByInventory(ctx, &rentalv1.ListRentalsByInventoryRequest{
			InventoryId: inventoryID,
			PageSize:    pageSize,
			Page:        page,
		})
	default:
		resp, err = h.rentalClient.ListRentals(ctx, &rentalv1.ListRentalsRequest{
			PageSize: pageSize,
			Page:     page,
		})
	}
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	rentals := make([]rentalResponse, len(resp.GetRentals()))
	for i, rental := range resp.GetRentals() {
		rentals[i] = rentalToResponse(rental)
	}

	writeJSON(w, http.StatusOK, rentalListResponse{
		Rentals:    rentals,
		TotalCount: resp.GetTotalCount(),
	})
}

// ListOverdueRentals returns a paginated list of overdue rentals.
func (h *RentalHandler) ListOverdueRentals(w http.ResponseWriter, r *http.Request) {
	pageSize, page := parsePagination(r)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.rentalClient.ListOverdueRentals(ctx, &rentalv1.ListOverdueRentalsRequest{
		PageSize: pageSize,
		Page:     page,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	rentals := make([]rentalResponse, len(resp.GetRentals()))
	for i, rental := range resp.GetRentals() {
		rentals[i] = rentalToResponse(rental)
	}

	writeJSON(w, http.StatusOK, rentalListResponse{
		Rentals:    rentals,
		TotalCount: resp.GetTotalCount(),
	})
}

// GetRental returns a single rental with full details.
func (h *RentalHandler) GetRental(w http.ResponseWriter, r *http.Request) {
	rentalID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid rental id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	detail, err := h.rentalClient.GetRental(ctx, &rentalv1.GetRentalRequest{
		RentalId: rentalID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, rentalDetailToResponse(detail))
}

// CreateRental creates a new rental.
func (h *RentalHandler) CreateRental(w http.ResponseWriter, r *http.Request) {
	var req createRentalRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rental, err := h.rentalClient.CreateRental(ctx, &rentalv1.CreateRentalRequest{
		InventoryId: req.InventoryID,
		CustomerId:  req.CustomerID,
		StaffId:     req.StaffID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, rentalToResponse(rental))
}

// ReturnRental marks a rental as returned.
func (h *RentalHandler) ReturnRental(w http.ResponseWriter, r *http.Request) {
	rentalID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid rental id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rental, err := h.rentalClient.ReturnRental(ctx, &rentalv1.ReturnRentalRequest{
		RentalId: rentalID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, rentalToResponse(rental))
}

// DeleteRental deletes a rental by ID.
func (h *RentalHandler) DeleteRental(w http.ResponseWriter, r *http.Request) {
	rentalID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid rental id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, err = h.rentalClient.DeleteRental(ctx, &rentalv1.DeleteRentalRequest{
		RentalId: rentalID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
