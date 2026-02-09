package handler

import (
	"context"
	"net/http"
	"time"

	rentalv1 "github.com/enkaigaku/dvd-rental/gen/proto/rental/v1"
	"github.com/enkaigaku/dvd-rental/pkg/middleware"
)

// RentalHandler handles rental endpoints (all require auth).
type RentalHandler struct {
	rentalClient rentalv1.RentalServiceClient
}

// NewRentalHandler creates a new RentalHandler.
func NewRentalHandler(rentalClient rentalv1.RentalServiceClient) *RentalHandler {
	return &RentalHandler{rentalClient: rentalClient}
}

// --- JSON models ---

type rentalItem struct {
	ID          int32  `json:"id"`
	InventoryID int32  `json:"inventory_id"`
	RentalDate  string `json:"rental_date"`
	ReturnDate  string `json:"return_date,omitempty"`
	Status      string `json:"status"`
}

type rentalDetailResponse struct {
	ID          int32  `json:"id"`
	CustomerID  int32  `json:"customer_id"`
	InventoryID int32  `json:"inventory_id"`
	StaffID     int32  `json:"staff_id"`
	RentalDate  string `json:"rental_date"`
	ReturnDate  string `json:"return_date,omitempty"`
	FilmTitle   string `json:"film_title,omitempty"`
	Status      string `json:"status"`
}

type rentalListResponse struct {
	Rentals    []rentalItem `json:"rentals"`
	TotalCount int32        `json:"total_count"`
	Page       int32        `json:"page"`
	PageSize   int32        `json:"page_size"`
}

type createRentalRequest struct {
	InventoryID int32 `json:"inventory_id"`
	StaffID     int32 `json:"staff_id"`
}

func rentalStatus(r *rentalv1.Rental) string {
	if r.GetReturnDate() != nil {
		return "returned"
	}
	return "active"
}

// ListRentals returns the authenticated customer's rentals.
func (h *RentalHandler) ListRentals(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	pageSize, page := parsePagination(r)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.rentalClient.ListRentalsByCustomer(ctx, &rentalv1.ListRentalsByCustomerRequest{
		CustomerId: claims.UserID,
		PageSize:   pageSize,
		Page:       page,
	})
	if err != nil {
		grpcToHTTPError(w, err)
		return
	}

	rentals := make([]rentalItem, len(resp.GetRentals()))
	for i, rental := range resp.GetRentals() {
		rentals[i] = rentalItem{
			ID:          rental.GetRentalId(),
			InventoryID: rental.GetInventoryId(),
			RentalDate:  timestampToString(rental.GetRentalDate()),
			ReturnDate:  timestampToString(rental.GetReturnDate()),
			Status:      rentalStatus(rental),
		}
	}

	middleware.WriteJSON(w, http.StatusOK, rentalListResponse{
		Rentals:    rentals,
		TotalCount: resp.GetTotalCount(),
		Page:       page,
		PageSize:   pageSize,
	})
}

// GetRental returns rental detail (verifies ownership).
func (h *RentalHandler) GetRental(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	rentalID, err := parseID(r, "id")
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	detail, err := h.rentalClient.GetRental(ctx, &rentalv1.GetRentalRequest{RentalId: rentalID})
	if err != nil {
		grpcToHTTPError(w, err)
		return
	}

	// Verify ownership.
	if detail.GetRental().GetCustomerId() != claims.UserID {
		middleware.WriteJSONError(w, http.StatusForbidden, "FORBIDDEN", "you can only access your own rentals")
		return
	}

	rental := detail.GetRental()
	middleware.WriteJSON(w, http.StatusOK, rentalDetailResponse{
		ID:          rental.GetRentalId(),
		CustomerID:  rental.GetCustomerId(),
		InventoryID: rental.GetInventoryId(),
		StaffID:     rental.GetStaffId(),
		RentalDate:  timestampToString(rental.GetRentalDate()),
		ReturnDate:  timestampToString(rental.GetReturnDate()),
		FilmTitle:   detail.GetFilmTitle(),
		Status:      rentalStatus(rental),
	})
}

// CreateRental creates a new rental for the authenticated customer.
func (h *RentalHandler) CreateRental(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var req createRentalRequest
	if err := readJSON(r, &req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	if req.InventoryID == 0 || req.StaffID == 0 {
		middleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "inventory_id and staff_id are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rental, err := h.rentalClient.CreateRental(ctx, &rentalv1.CreateRentalRequest{
		CustomerId:  claims.UserID,
		InventoryId: req.InventoryID,
		StaffId:     req.StaffID,
	})
	if err != nil {
		grpcToHTTPError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusCreated, rentalItem{
		ID:          rental.GetRentalId(),
		InventoryID: rental.GetInventoryId(),
		RentalDate:  timestampToString(rental.GetRentalDate()),
		Status:      "active",
	})
}

// ReturnRental marks a rental as returned (verifies ownership).
func (h *RentalHandler) ReturnRental(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	rentalID, err := parseID(r, "id")
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Verify ownership first.
	detail, err := h.rentalClient.GetRental(ctx, &rentalv1.GetRentalRequest{RentalId: rentalID})
	if err != nil {
		grpcToHTTPError(w, err)
		return
	}
	if detail.GetRental().GetCustomerId() != claims.UserID {
		middleware.WriteJSONError(w, http.StatusForbidden, "FORBIDDEN", "you can only return your own rentals")
		return
	}

	// Return the rental.
	rental, err := h.rentalClient.ReturnRental(ctx, &rentalv1.ReturnRentalRequest{RentalId: rentalID})
	if err != nil {
		grpcToHTTPError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, rentalItem{
		ID:          rental.GetRentalId(),
		InventoryID: rental.GetInventoryId(),
		RentalDate:  timestampToString(rental.GetRentalDate()),
		ReturnDate:  timestampToString(rental.GetReturnDate()),
		Status:      "returned",
	})
}
