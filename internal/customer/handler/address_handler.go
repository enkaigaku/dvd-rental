package handler

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	customerv1 "github.com/enkaigaku/dvd-rental/gen/proto/customer/v1"
	"github.com/enkaigaku/dvd-rental/internal/customer/model"
	"github.com/enkaigaku/dvd-rental/internal/customer/repository"
	"github.com/enkaigaku/dvd-rental/internal/customer/service"
)

// AddressHandler implements the AddressService gRPC server.
type AddressHandler struct {
	customerv1.UnimplementedAddressServiceServer
	svc *service.AddressService
}

// NewAddressHandler creates a new AddressHandler.
func NewAddressHandler(svc *service.AddressService) *AddressHandler {
	return &AddressHandler{svc: svc}
}

func (h *AddressHandler) GetAddress(ctx context.Context, req *customerv1.GetAddressRequest) (*customerv1.Address, error) {
	addr, err := h.svc.GetAddress(ctx, req.GetAddressId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return addressToProto(addr), nil
}

func (h *AddressHandler) ListAddresses(ctx context.Context, req *customerv1.ListAddressesRequest) (*customerv1.ListAddressesResponse, error) {
	addresses, total, err := h.svc.ListAddresses(ctx, req.GetPageSize(), req.GetPage())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toAddressListResponse(addresses, total), nil
}

func (h *AddressHandler) CreateAddress(ctx context.Context, req *customerv1.CreateAddressRequest) (*customerv1.Address, error) {
	addr, err := h.svc.CreateAddress(ctx, repository.CreateAddressParams{
		Address:    req.GetAddress(),
		Address2:   req.GetAddress2(),
		District:   req.GetDistrict(),
		CityID:     req.GetCityId(),
		PostalCode: req.GetPostalCode(),
		Phone:      req.GetPhone(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return addressToProto(addr), nil
}

func (h *AddressHandler) UpdateAddress(ctx context.Context, req *customerv1.UpdateAddressRequest) (*customerv1.Address, error) {
	addr, err := h.svc.UpdateAddress(ctx, repository.UpdateAddressParams{
		AddressID:  req.GetAddressId(),
		Address:    req.GetAddress(),
		Address2:   req.GetAddress2(),
		District:   req.GetDistrict(),
		CityID:     req.GetCityId(),
		PostalCode: req.GetPostalCode(),
		Phone:      req.GetPhone(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return addressToProto(addr), nil
}

func (h *AddressHandler) DeleteAddress(ctx context.Context, req *customerv1.DeleteAddressRequest) (*emptypb.Empty, error) {
	if err := h.svc.DeleteAddress(ctx, req.GetAddressId()); err != nil {
		return nil, toGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

func toAddressListResponse(addresses []model.Address, total int64) *customerv1.ListAddressesResponse {
	protos := make([]*customerv1.Address, len(addresses))
	for i, a := range addresses {
		protos[i] = addressToProto(a)
	}
	return &customerv1.ListAddressesResponse{
		Addresses:  protos,
		TotalCount: int32(total),
	}
}
