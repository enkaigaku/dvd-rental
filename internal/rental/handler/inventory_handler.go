package handler

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	rentalv1 "github.com/enkaigaku/dvd-rental/gen/proto/rental/v1"
	"github.com/enkaigaku/dvd-rental/internal/rental/model"
	"github.com/enkaigaku/dvd-rental/internal/rental/repository"
	"github.com/enkaigaku/dvd-rental/internal/rental/service"
)

// InventoryHandler implements the InventoryService gRPC server.
type InventoryHandler struct {
	rentalv1.UnimplementedInventoryServiceServer
	svc *service.InventoryService
}

// NewInventoryHandler creates a new InventoryHandler.
func NewInventoryHandler(svc *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{svc: svc}
}

func (h *InventoryHandler) GetInventory(ctx context.Context, req *rentalv1.GetInventoryRequest) (*rentalv1.Inventory, error) {
	inv, err := h.svc.GetInventory(ctx, req.GetInventoryId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return inventoryToProto(inv), nil
}

func (h *InventoryHandler) ListInventory(ctx context.Context, req *rentalv1.ListInventoryRequest) (*rentalv1.ListInventoryResponse, error) {
	items, total, err := h.svc.ListInventory(ctx, req.GetPageSize(), req.GetPage())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toInventoryListResponse(items, total), nil
}

func (h *InventoryHandler) ListInventoryByFilm(ctx context.Context, req *rentalv1.ListInventoryByFilmRequest) (*rentalv1.ListInventoryResponse, error) {
	items, total, err := h.svc.ListInventoryByFilm(ctx, req.GetFilmId(), req.GetPageSize(), req.GetPage())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toInventoryListResponse(items, total), nil
}

func (h *InventoryHandler) ListInventoryByStore(ctx context.Context, req *rentalv1.ListInventoryByStoreRequest) (*rentalv1.ListInventoryResponse, error) {
	items, total, err := h.svc.ListInventoryByStore(ctx, req.GetStoreId(), req.GetPageSize(), req.GetPage())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toInventoryListResponse(items, total), nil
}

func (h *InventoryHandler) CheckInventoryAvailability(ctx context.Context, req *rentalv1.CheckInventoryAvailabilityRequest) (*rentalv1.CheckInventoryAvailabilityResponse, error) {
	available, err := h.svc.CheckInventoryAvailability(ctx, req.GetInventoryId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &rentalv1.CheckInventoryAvailabilityResponse{Available: available}, nil
}

func (h *InventoryHandler) ListAvailableInventory(ctx context.Context, req *rentalv1.ListAvailableInventoryRequest) (*rentalv1.ListInventoryResponse, error) {
	items, total, err := h.svc.ListAvailableInventory(ctx, req.GetFilmId(), req.GetStoreId(), req.GetPageSize(), req.GetPage())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toInventoryListResponse(items, total), nil
}

func (h *InventoryHandler) CreateInventory(ctx context.Context, req *rentalv1.CreateInventoryRequest) (*rentalv1.Inventory, error) {
	inv, err := h.svc.CreateInventory(ctx, repository.CreateInventoryParams{
		FilmID:  req.GetFilmId(),
		StoreID: req.GetStoreId(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return inventoryToProto(inv), nil
}

func (h *InventoryHandler) DeleteInventory(ctx context.Context, req *rentalv1.DeleteInventoryRequest) (*emptypb.Empty, error) {
	if err := h.svc.DeleteInventory(ctx, req.GetInventoryId()); err != nil {
		return nil, toGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

func toInventoryListResponse(items []model.Inventory, total int64) *rentalv1.ListInventoryResponse {
	protos := make([]*rentalv1.Inventory, len(items))
	for i, item := range items {
		protos[i] = inventoryToProto(item)
	}
	return &rentalv1.ListInventoryResponse{
		Items:      protos,
		TotalCount: int32(total),
	}
}
