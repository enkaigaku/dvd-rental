package handler

import (
	"context"
	"net/http"
	"time"

	storev1 "github.com/tokyoyuan/dvd-rental/gen/proto/store/v1"
)

// StaffHandler handles staff management endpoints.
type StaffHandler struct {
	staffClient storev1.StaffServiceClient
}

// NewStaffHandler creates a new StaffHandler.
func NewStaffHandler(staffClient storev1.StaffServiceClient) *StaffHandler {
	return &StaffHandler{staffClient: staffClient}
}

// --- JSON models ---

type staffResponse struct {
	StaffID    int32  `json:"staff_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	AddressID  int32  `json:"address_id"`
	Email      string `json:"email"`
	StoreID    int32  `json:"store_id"`
	Active     bool   `json:"active"`
	Username   string `json:"username"`
	LastUpdate string `json:"last_update"`
}

type staffListResponse struct {
	Staff      []staffResponse `json:"staff"`
	TotalCount int32           `json:"total_count"`
}

type createStaffRequest struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	AddressID    int32  `json:"address_id"`
	Email        string `json:"email"`
	StoreID      int32  `json:"store_id"`
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
}

type updateStaffRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AddressID int32  `json:"address_id"`
	Email     string `json:"email"`
	StoreID   int32  `json:"store_id"`
	Username  string `json:"username"`
}

type updatePasswordRequest struct {
	PasswordHash string `json:"password_hash"`
}

func staffToResponse(s *storev1.Staff) staffResponse {
	return staffResponse{
		StaffID:    s.GetStaffId(),
		FirstName:  s.GetFirstName(),
		LastName:   s.GetLastName(),
		AddressID:  s.GetAddressId(),
		Email:      s.GetEmail(),
		StoreID:    s.GetStoreId(),
		Active:     s.GetActive(),
		Username:   s.GetUsername(),
		LastUpdate: s.GetLastUpdate().AsTime().Format(time.RFC3339),
	}
}

// ListStaff returns a paginated list of staff members.
func (h *StaffHandler) ListStaff(w http.ResponseWriter, r *http.Request) {
	pageSize, page := parsePagination(r)
	storeID := parseQueryInt32(r, "store_id")
	activeOnly := parseQueryBool(r, "active_only")

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var resp *storev1.ListStaffResponse
	var err error

	if storeID > 0 {
		resp, err = h.staffClient.ListStaffByStore(ctx, &storev1.ListStaffByStoreRequest{
			StoreId:    storeID,
			PageSize:   pageSize,
			Page:       page,
			ActiveOnly: activeOnly,
		})
	} else {
		resp, err = h.staffClient.ListStaff(ctx, &storev1.ListStaffRequest{
			PageSize:   pageSize,
			Page:       page,
			ActiveOnly: activeOnly,
		})
	}
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	staff := make([]staffResponse, len(resp.GetStaff()))
	for i, s := range resp.GetStaff() {
		staff[i] = staffToResponse(s)
	}

	writeJSON(w, http.StatusOK, staffListResponse{
		Staff:      staff,
		TotalCount: resp.GetTotalCount(),
	})
}

// GetStaff returns a single staff member by ID.
func (h *StaffHandler) GetStaff(w http.ResponseWriter, r *http.Request) {
	staffID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid staff id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	staff, err := h.staffClient.GetStaff(ctx, &storev1.GetStaffRequest{
		StaffId: staffID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, staffToResponse(staff))
}

// CreateStaff creates a new staff member.
func (h *StaffHandler) CreateStaff(w http.ResponseWriter, r *http.Request) {
	var req createStaffRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	staff, err := h.staffClient.CreateStaff(ctx, &storev1.CreateStaffRequest{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		AddressId:    req.AddressID,
		Email:        req.Email,
		StoreId:      req.StoreID,
		Username:     req.Username,
		PasswordHash: req.PasswordHash,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, staffToResponse(staff))
}

// UpdateStaff updates an existing staff member.
func (h *StaffHandler) UpdateStaff(w http.ResponseWriter, r *http.Request) {
	staffID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid staff id")
		return
	}

	var req updateStaffRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	staff, err := h.staffClient.UpdateStaff(ctx, &storev1.UpdateStaffRequest{
		StaffId:   staffID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		AddressId: req.AddressID,
		Email:     req.Email,
		StoreId:   req.StoreID,
		Username:  req.Username,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, staffToResponse(staff))
}

// DeactivateStaff deactivates a staff member (soft delete).
func (h *StaffHandler) DeactivateStaff(w http.ResponseWriter, r *http.Request) {
	staffID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid staff id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	staff, err := h.staffClient.DeactivateStaff(ctx, &storev1.DeactivateStaffRequest{
		StaffId: staffID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, staffToResponse(staff))
}

// UpdateStaffPassword updates a staff member's password.
func (h *StaffHandler) UpdateStaffPassword(w http.ResponseWriter, r *http.Request) {
	staffID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid staff id")
		return
	}

	var req updatePasswordRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	if req.PasswordHash == "" {
		writeError(w, http.StatusBadRequest, "password_hash is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, err = h.staffClient.UpdateStaffPassword(ctx, &storev1.UpdateStaffPasswordRequest{
		StaffId:      staffID,
		PasswordHash: req.PasswordHash,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
