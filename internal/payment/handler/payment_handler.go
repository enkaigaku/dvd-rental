package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	paymentv1 "github.com/tokyoyuan/dvd-rental/gen/proto/payment/v1"
	"github.com/tokyoyuan/dvd-rental/internal/payment/model"
	"github.com/tokyoyuan/dvd-rental/internal/payment/repository"
	"github.com/tokyoyuan/dvd-rental/internal/payment/service"
)

// PaymentHandler implements the PaymentService gRPC server.
type PaymentHandler struct {
	paymentv1.UnimplementedPaymentServiceServer
	svc *service.PaymentService
}

// NewPaymentHandler creates a new PaymentHandler.
func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

func (h *PaymentHandler) GetPayment(ctx context.Context, req *paymentv1.GetPaymentRequest) (*paymentv1.PaymentDetail, error) {
	detail, err := h.svc.GetPayment(ctx, req.GetPaymentId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return paymentDetailToProto(detail), nil
}

func (h *PaymentHandler) ListPayments(ctx context.Context, req *paymentv1.ListPaymentsRequest) (*paymentv1.ListPaymentsResponse, error) {
	payments, total, err := h.svc.ListPayments(ctx, req.GetPageSize(), req.GetPage())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toPaymentListResponse(payments, total), nil
}

func (h *PaymentHandler) ListPaymentsByCustomer(ctx context.Context, req *paymentv1.ListPaymentsByCustomerRequest) (*paymentv1.ListPaymentsResponse, error) {
	payments, total, err := h.svc.ListPaymentsByCustomer(ctx, req.GetCustomerId(), req.GetPageSize(), req.GetPage())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toPaymentListResponse(payments, total), nil
}

func (h *PaymentHandler) ListPaymentsByStaff(ctx context.Context, req *paymentv1.ListPaymentsByStaffRequest) (*paymentv1.ListPaymentsResponse, error) {
	payments, total, err := h.svc.ListPaymentsByStaff(ctx, req.GetStaffId(), req.GetPageSize(), req.GetPage())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toPaymentListResponse(payments, total), nil
}

func (h *PaymentHandler) ListPaymentsByRental(ctx context.Context, req *paymentv1.ListPaymentsByRentalRequest) (*paymentv1.ListPaymentsResponse, error) {
	payments, total, err := h.svc.ListPaymentsByRental(ctx, req.GetRentalId(), req.GetPageSize(), req.GetPage())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toPaymentListResponse(payments, total), nil
}

func (h *PaymentHandler) ListPaymentsByDateRange(ctx context.Context, req *paymentv1.ListPaymentsByDateRangeRequest) (*paymentv1.ListPaymentsResponse, error) {
	if req.GetStartDate() == nil {
		return nil, status.Error(codes.InvalidArgument, "start_date is required")
	}
	if req.GetEndDate() == nil {
		return nil, status.Error(codes.InvalidArgument, "end_date is required")
	}

	startDate := req.GetStartDate().AsTime()
	endDate := req.GetEndDate().AsTime()

	payments, total, err := h.svc.ListPaymentsByDateRange(ctx, startDate, endDate, req.GetPageSize(), req.GetPage())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toPaymentListResponse(payments, total), nil
}

func (h *PaymentHandler) CreatePayment(ctx context.Context, req *paymentv1.CreatePaymentRequest) (*paymentv1.Payment, error) {
	payment, err := h.svc.CreatePayment(ctx, repository.CreatePaymentParams{
		CustomerID: req.GetCustomerId(),
		StaffID:    req.GetStaffId(),
		RentalID:   req.GetRentalId(),
		Amount:     req.GetAmount(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return paymentToProto(payment), nil
}

func (h *PaymentHandler) DeletePayment(ctx context.Context, req *paymentv1.DeletePaymentRequest) (*emptypb.Empty, error) {
	if err := h.svc.DeletePayment(ctx, req.GetPaymentId()); err != nil {
		return nil, toGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

func toPaymentListResponse(payments []model.Payment, total int64) *paymentv1.ListPaymentsResponse {
	protos := make([]*paymentv1.Payment, len(payments))
	for i, p := range payments {
		protos[i] = paymentToProto(p)
	}
	return &paymentv1.ListPaymentsResponse{
		Payments:   protos,
		TotalCount: int32(total),
	}
}
