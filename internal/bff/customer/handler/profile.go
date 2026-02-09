package handler

import (
	"context"
	"net/http"
	"time"

	customerv1 "github.com/enkaigaku/dvd-rental/gen/proto/customer/v1"
	"github.com/enkaigaku/dvd-rental/pkg/middleware"
)

// ProfileHandler handles customer profile endpoints (all require auth).
type ProfileHandler struct {
	customerClient customerv1.CustomerServiceClient
}

// NewProfileHandler creates a new ProfileHandler.
func NewProfileHandler(customerClient customerv1.CustomerServiceClient) *ProfileHandler {
	return &ProfileHandler{customerClient: customerClient}
}

// --- JSON models ---

type profileResponse struct {
	ID         int32  `json:"id"`
	StoreID    int32  `json:"store_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	Active     bool   `json:"active"`
	CreateDate string `json:"create_date"`
	Address    string `json:"address,omitempty"`
	District   string `json:"district,omitempty"`
	City       string `json:"city,omitempty"`
	Country    string `json:"country,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	Phone      string `json:"phone,omitempty"`
}

type updateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

// GetProfile returns the authenticated customer's profile.
func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	detail, err := h.customerClient.GetCustomer(ctx, &customerv1.GetCustomerRequest{
		CustomerId: claims.UserID,
	})
	if err != nil {
		grpcToHTTPError(w, err)
		return
	}

	c := detail.GetCustomer()
	middleware.WriteJSON(w, http.StatusOK, profileResponse{
		ID:         c.GetCustomerId(),
		StoreID:    c.GetStoreId(),
		FirstName:  c.GetFirstName(),
		LastName:   c.GetLastName(),
		Email:      c.GetEmail(),
		Active:     c.GetActive(),
		CreateDate: timestampToString(c.GetCreateDate()),
		Address:    detail.GetAddress(),
		District:   detail.GetDistrict(),
		City:       detail.GetCity(),
		Country:    detail.GetCountry(),
		PostalCode: detail.GetPostalCode(),
		Phone:      detail.GetPhone(),
	})
}

// UpdateProfile updates the authenticated customer's profile.
// Uses read-modify-write since UpdateCustomerRequest requires all fields.
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var req updateProfileRequest
	if err := readJSON(r, &req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	if req.FirstName == "" || req.LastName == "" || req.Email == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "first_name, last_name, and email are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Read current customer to get fields we don't allow changing.
	detail, err := h.customerClient.GetCustomer(ctx, &customerv1.GetCustomerRequest{
		CustomerId: claims.UserID,
	})
	if err != nil {
		grpcToHTTPError(w, err)
		return
	}

	current := detail.GetCustomer()

	// Merge: user-submitted fields + preserved fields.
	updated, err := h.customerClient.UpdateCustomer(ctx, &customerv1.UpdateCustomerRequest{
		CustomerId: claims.UserID,
		StoreId:    current.GetStoreId(),
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Email:      req.Email,
		AddressId:  current.GetAddressId(),
		Active:     current.GetActive(),
	})
	if err != nil {
		grpcToHTTPError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, profileResponse{
		ID:        updated.GetCustomerId(),
		StoreID:   updated.GetStoreId(),
		FirstName: updated.GetFirstName(),
		LastName:  updated.GetLastName(),
		Email:     updated.GetEmail(),
		Active:    updated.GetActive(),
	})
}
