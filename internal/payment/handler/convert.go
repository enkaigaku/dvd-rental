package handler

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	paymentv1 "github.com/tokyoyuan/dvd-rental/gen/proto/payment/v1"
	"github.com/tokyoyuan/dvd-rental/internal/payment/model"
	"github.com/tokyoyuan/dvd-rental/internal/payment/service"
)

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, service.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, service.ErrForeignKey):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func paymentToProto(p model.Payment) *paymentv1.Payment {
	return &paymentv1.Payment{
		PaymentId:   p.PaymentID,
		CustomerId:  p.CustomerID,
		StaffId:     p.StaffID,
		RentalId:    p.RentalID,
		Amount:      p.Amount,
		PaymentDate: timestamppb.New(p.PaymentDate),
	}
}

func paymentDetailToProto(d model.PaymentDetail) *paymentv1.PaymentDetail {
	pb := &paymentv1.PaymentDetail{
		Payment:      paymentToProto(d.Payment),
		CustomerName: d.CustomerName,
		StaffName:    d.StaffName,
	}
	// Zero time means rental not found — leave rental_date nil in proto.
	if !d.RentalDate.IsZero() {
		pb.RentalDate = timestamppb.New(d.RentalDate)
	}
	return pb
}
