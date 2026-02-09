package handler

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	customerv1 "github.com/tokyoyuan/dvd-rental/gen/proto/customer/v1"
	"github.com/tokyoyuan/dvd-rental/internal/customer/model"
	"github.com/tokyoyuan/dvd-rental/internal/customer/repository"
	"github.com/tokyoyuan/dvd-rental/internal/customer/service"
)

// CustomerHandler implements the CustomerService gRPC server.
type CustomerHandler struct {
	customerv1.UnimplementedCustomerServiceServer
	svc *service.CustomerService
}

// NewCustomerHandler creates a new CustomerHandler.
func NewCustomerHandler(svc *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{svc: svc}
}

func (h *CustomerHandler) GetCustomer(ctx context.Context, req *customerv1.GetCustomerRequest) (*customerv1.CustomerDetail, error) {
	detail, err := h.svc.GetCustomer(ctx, req.GetCustomerId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return customerDetailToProto(detail), nil
}

func (h *CustomerHandler) ListCustomers(ctx context.Context, req *customerv1.ListCustomersRequest) (*customerv1.ListCustomersResponse, error) {
	customers, total, err := h.svc.ListCustomers(ctx, req.GetPageSize(), req.GetPage())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toCustomerListResponse(customers, total), nil
}

func (h *CustomerHandler) ListCustomersByStore(ctx context.Context, req *customerv1.ListCustomersByStoreRequest) (*customerv1.ListCustomersResponse, error) {
	customers, total, err := h.svc.ListCustomersByStore(ctx, req.GetStoreId(), req.GetPageSize(), req.GetPage())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toCustomerListResponse(customers, total), nil
}

func (h *CustomerHandler) CreateCustomer(ctx context.Context, req *customerv1.CreateCustomerRequest) (*customerv1.Customer, error) {
	cust, err := h.svc.CreateCustomer(ctx, repository.CreateCustomerParams{
		StoreID:   req.GetStoreId(),
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
		Email:     req.GetEmail(),
		AddressID: req.GetAddressId(),
		Active:    req.GetActive(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return customerToProto(cust), nil
}

func (h *CustomerHandler) UpdateCustomer(ctx context.Context, req *customerv1.UpdateCustomerRequest) (*customerv1.Customer, error) {
	cust, err := h.svc.UpdateCustomer(ctx, repository.UpdateCustomerParams{
		CustomerID: req.GetCustomerId(),
		StoreID:    req.GetStoreId(),
		FirstName:  req.GetFirstName(),
		LastName:   req.GetLastName(),
		Email:      req.GetEmail(),
		AddressID:  req.GetAddressId(),
		Active:     req.GetActive(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return customerToProto(cust), nil
}

func (h *CustomerHandler) DeleteCustomer(ctx context.Context, req *customerv1.DeleteCustomerRequest) (*emptypb.Empty, error) {
	if err := h.svc.DeleteCustomer(ctx, req.GetCustomerId()); err != nil {
		return nil, toGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

func toCustomerListResponse(customers []model.Customer, total int64) *customerv1.ListCustomersResponse {
	protos := make([]*customerv1.Customer, len(customers))
	for i, c := range customers {
		protos[i] = customerToProto(c)
	}
	return &customerv1.ListCustomersResponse{
		Customers:  protos,
		TotalCount: int32(total),
	}
}
