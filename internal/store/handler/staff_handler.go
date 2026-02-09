package handler

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	storev1 "github.com/enkaigaku/dvd-rental/gen/proto/store/v1"
	"github.com/enkaigaku/dvd-rental/internal/store/model"
	"github.com/enkaigaku/dvd-rental/internal/store/repository"
	"github.com/enkaigaku/dvd-rental/internal/store/service"
)

// StaffHandler implements the StaffServiceServer gRPC interface.
type StaffHandler struct {
	storev1.UnimplementedStaffServiceServer
	svc *service.StaffService
}

// NewStaffHandler creates a new StaffHandler.
func NewStaffHandler(svc *service.StaffService) *StaffHandler {
	return &StaffHandler{svc: svc}
}

// GetStaff retrieves a staff member by ID.
func (h *StaffHandler) GetStaff(ctx context.Context, req *storev1.GetStaffRequest) (*storev1.Staff, error) {
	staff, err := h.svc.GetStaff(ctx, req.GetStaffId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return staffToProto(staff), nil
}

// GetStaffByUsername retrieves a staff member by username (includes password_hash for BFF auth).
func (h *StaffHandler) GetStaffByUsername(ctx context.Context, req *storev1.GetStaffByUsernameRequest) (*storev1.Staff, error) {
	staff, err := h.svc.GetStaffByUsername(ctx, req.GetUsername())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return staffToProto(staff), nil
}

// ListStaff retrieves a paginated list of all staff.
func (h *StaffHandler) ListStaff(ctx context.Context, req *storev1.ListStaffRequest) (*storev1.ListStaffResponse, error) {
	staff, total, err := h.svc.ListStaff(ctx, req.GetPageSize(), req.GetPage(), req.GetActiveOnly())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toStaffListResponse(staff, total), nil
}

// ListStaffByStore retrieves a paginated list of staff for a specific store.
func (h *StaffHandler) ListStaffByStore(ctx context.Context, req *storev1.ListStaffByStoreRequest) (*storev1.ListStaffResponse, error) {
	staff, total, err := h.svc.ListStaffByStore(ctx, req.GetStoreId(), req.GetPageSize(), req.GetPage(), req.GetActiveOnly())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toStaffListResponse(staff, total), nil
}

// CreateStaff creates a new staff member.
func (h *StaffHandler) CreateStaff(ctx context.Context, req *storev1.CreateStaffRequest) (*storev1.Staff, error) {
	staff, err := h.svc.CreateStaff(ctx, repository.CreateStaffParams{
		FirstName:    req.GetFirstName(),
		LastName:     req.GetLastName(),
		AddressID:    req.GetAddressId(),
		Email:        req.GetEmail(),
		StoreID:      req.GetStoreId(),
		Username:     req.GetUsername(),
		PasswordHash: req.GetPasswordHash(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return staffToProto(staff), nil
}

// UpdateStaff updates a staff member's information.
func (h *StaffHandler) UpdateStaff(ctx context.Context, req *storev1.UpdateStaffRequest) (*storev1.Staff, error) {
	staff, err := h.svc.UpdateStaff(ctx, repository.UpdateStaffParams{
		StaffID:   req.GetStaffId(),
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
		AddressID: req.GetAddressId(),
		Email:     req.GetEmail(),
		StoreID:   req.GetStoreId(),
		Username:  req.GetUsername(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return staffToProto(staff), nil
}

// DeactivateStaff deactivates a staff member (soft delete).
func (h *StaffHandler) DeactivateStaff(ctx context.Context, req *storev1.DeactivateStaffRequest) (*storev1.Staff, error) {
	staff, err := h.svc.DeactivateStaff(ctx, req.GetStaffId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return staffToProto(staff), nil
}

// UpdateStaffPassword updates a staff member's password hash.
func (h *StaffHandler) UpdateStaffPassword(ctx context.Context, req *storev1.UpdateStaffPasswordRequest) (*emptypb.Empty, error) {
	if err := h.svc.UpdateStaffPassword(ctx, req.GetStaffId(), req.GetPasswordHash()); err != nil {
		return nil, toGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

func toStaffListResponse(staff []model.Staff, total int64) *storev1.ListStaffResponse {
	pbStaff := make([]*storev1.Staff, len(staff))
	for i, s := range staff {
		pbStaff[i] = staffToProto(s)
	}
	return &storev1.ListStaffResponse{
		Staff:      pbStaff,
		TotalCount: int32(total),
	}
}
