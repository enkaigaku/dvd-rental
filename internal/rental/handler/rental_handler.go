package handler

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	rentalv1 "github.com/enkaigaku/dvd-rental/gen/proto/rental/v1"
	"github.com/enkaigaku/dvd-rental/internal/rental/model"
	"github.com/enkaigaku/dvd-rental/internal/rental/repository"
	"github.com/enkaigaku/dvd-rental/internal/rental/service"
)

// RentalHandler implements the RentalService gRPC server.
type RentalHandler struct {
	rentalv1.UnimplementedRentalServiceServer
	svc *service.RentalService
}

// NewRentalHandler creates a new RentalHandler.
func NewRentalHandler(svc *service.RentalService) *RentalHandler {
	return &RentalHandler{svc: svc}
}

func (h *RentalHandler) GetRental(ctx context.Context, req *rentalv1.GetRentalRequest) (*rentalv1.RentalDetail, error) {
	detail, err := h.svc.GetRental(ctx, req.GetRentalId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return rentalDetailToProto(detail), nil
}

func (h *RentalHandler) ListRentals(ctx context.Context, req *rentalv1.ListRentalsRequest) (*rentalv1.ListRentalsResponse, error) {
	rentals, total, err := h.svc.ListRentals(ctx, req.GetPageSize(), req.GetPage())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toRentalListResponse(rentals, total), nil
}

func (h *RentalHandler) ListRentalsByCustomer(ctx context.Context, req *rentalv1.ListRentalsByCustomerRequest) (*rentalv1.ListRentalsResponse, error) {
	rentals, total, err := h.svc.ListRentalsByCustomer(ctx, req.GetCustomerId(), req.GetPageSize(), req.GetPage())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toRentalListResponse(rentals, total), nil
}

func (h *RentalHandler) ListRentalsByInventory(ctx context.Context, req *rentalv1.ListRentalsByInventoryRequest) (*rentalv1.ListRentalsResponse, error) {
	rentals, total, err := h.svc.ListRentalsByInventory(ctx, req.GetInventoryId(), req.GetPageSize(), req.GetPage())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toRentalListResponse(rentals, total), nil
}

func (h *RentalHandler) ListOverdueRentals(ctx context.Context, req *rentalv1.ListOverdueRentalsRequest) (*rentalv1.ListRentalsResponse, error) {
	rentals, total, err := h.svc.ListOverdueRentals(ctx, req.GetPageSize(), req.GetPage())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toRentalListResponse(rentals, total), nil
}

func (h *RentalHandler) CreateRental(ctx context.Context, req *rentalv1.CreateRentalRequest) (*rentalv1.Rental, error) {
	rental, err := h.svc.CreateRental(ctx, repository.CreateRentalParams{
		InventoryID: req.GetInventoryId(),
		CustomerID:  req.GetCustomerId(),
		StaffID:     req.GetStaffId(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return rentalToProto(rental), nil
}

func (h *RentalHandler) ReturnRental(ctx context.Context, req *rentalv1.ReturnRentalRequest) (*rentalv1.Rental, error) {
	rental, err := h.svc.ReturnRental(ctx, req.GetRentalId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return rentalToProto(rental), nil
}

func (h *RentalHandler) DeleteRental(ctx context.Context, req *rentalv1.DeleteRentalRequest) (*emptypb.Empty, error) {
	if err := h.svc.DeleteRental(ctx, req.GetRentalId()); err != nil {
		return nil, toGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

func toRentalListResponse(rentals []model.Rental, total int64) *rentalv1.ListRentalsResponse {
	protos := make([]*rentalv1.Rental, len(rentals))
	for i, r := range rentals {
		protos[i] = rentalToProto(r)
	}
	return &rentalv1.ListRentalsResponse{
		Rentals:    protos,
		TotalCount: int32(total),
	}
}
