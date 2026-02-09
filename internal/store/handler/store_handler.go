package handler

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	storev1 "github.com/enkaigaku/dvd-rental/gen/proto/store/v1"
	"github.com/enkaigaku/dvd-rental/internal/store/service"
)

// StoreHandler implements the StoreServiceServer gRPC interface.
type StoreHandler struct {
	storev1.UnimplementedStoreServiceServer
	svc *service.StoreService
}

// NewStoreHandler creates a new StoreHandler.
func NewStoreHandler(svc *service.StoreService) *StoreHandler {
	return &StoreHandler{svc: svc}
}

// GetStore retrieves a store by ID.
func (h *StoreHandler) GetStore(ctx context.Context, req *storev1.GetStoreRequest) (*storev1.Store, error) {
	store, err := h.svc.GetStore(ctx, req.GetStoreId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return storeToProto(store), nil
}

// ListStores retrieves a paginated list of stores.
func (h *StoreHandler) ListStores(ctx context.Context, req *storev1.ListStoresRequest) (*storev1.ListStoresResponse, error) {
	stores, total, err := h.svc.ListStores(ctx, req.GetPageSize(), req.GetPage())
	if err != nil {
		return nil, toGRPCError(err)
	}
	pbStores := make([]*storev1.Store, len(stores))
	for i, s := range stores {
		pbStores[i] = storeToProto(s)
	}
	return &storev1.ListStoresResponse{
		Stores:     pbStores,
		TotalCount: int32(total),
	}, nil
}

// CreateStore creates a new store.
func (h *StoreHandler) CreateStore(ctx context.Context, req *storev1.CreateStoreRequest) (*storev1.Store, error) {
	store, err := h.svc.CreateStore(ctx, req.GetManagerStaffId(), req.GetAddressId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return storeToProto(store), nil
}

// UpdateStore updates an existing store.
func (h *StoreHandler) UpdateStore(ctx context.Context, req *storev1.UpdateStoreRequest) (*storev1.Store, error) {
	store, err := h.svc.UpdateStore(ctx, req.GetStoreId(), req.GetManagerStaffId(), req.GetAddressId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return storeToProto(store), nil
}

// DeleteStore deletes a store by ID.
func (h *StoreHandler) DeleteStore(ctx context.Context, req *storev1.DeleteStoreRequest) (*emptypb.Empty, error) {
	if err := h.svc.DeleteStore(ctx, req.GetStoreId()); err != nil {
		return nil, toGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}
