package handler

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	rentalv1 "github.com/enkaigaku/dvd-rental/gen/proto/rental/v1"
	"github.com/enkaigaku/dvd-rental/internal/rental/model"
	"github.com/enkaigaku/dvd-rental/internal/rental/service"
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

func rentalToProto(r model.Rental) *rentalv1.Rental {
	pb := &rentalv1.Rental{
		RentalId:    r.RentalID,
		RentalDate:  timestamppb.New(r.RentalDate),
		InventoryId: r.InventoryID,
		CustomerId:  r.CustomerID,
		StaffId:     r.StaffID,
		LastUpdate:  timestamppb.New(r.LastUpdate),
	}
	// Zero time means not yet returned — leave return_date nil in proto.
	if !r.ReturnDate.IsZero() {
		pb.ReturnDate = timestamppb.New(r.ReturnDate)
	}
	return pb
}

func rentalDetailToProto(d model.RentalDetail) *rentalv1.RentalDetail {
	return &rentalv1.RentalDetail{
		Rental:       rentalToProto(d.Rental),
		CustomerName: d.CustomerName,
		FilmTitle:    d.FilmTitle,
		StoreId:      d.StoreID,
	}
}

func inventoryToProto(i model.Inventory) *rentalv1.Inventory {
	return &rentalv1.Inventory{
		InventoryId: i.InventoryID,
		FilmId:      i.FilmID,
		StoreId:     i.StoreID,
		LastUpdate:  timestamppb.New(i.LastUpdate),
	}
}
