// Package handler implements gRPC handlers for the store service.
package handler

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	storev1 "github.com/enkaigaku/dvd-rental/gen/proto/store/v1"
	"github.com/enkaigaku/dvd-rental/internal/store/model"
	"github.com/enkaigaku/dvd-rental/internal/store/service"
)

// toGRPCError maps service-layer sentinel errors to gRPC status errors.
func toGRPCError(err error) error {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, service.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

// storeToProto converts a domain Store to its protobuf representation.
func storeToProto(s model.Store) *storev1.Store {
	return &storev1.Store{
		StoreId:        s.StoreID,
		ManagerStaffId: s.ManagerStaffID,
		AddressId:      s.AddressID,
		LastUpdate:     timestamppb.New(s.LastUpdate),
	}
}

// staffToProto converts a domain Staff to its protobuf representation.
func staffToProto(s model.Staff) *storev1.Staff {
	return &storev1.Staff{
		StaffId:      s.StaffID,
		FirstName:    s.FirstName,
		LastName:     s.LastName,
		AddressId:    s.AddressID,
		Email:        s.Email,
		StoreId:      s.StoreID,
		Active:       s.Active,
		Username:     s.Username,
		PasswordHash: s.PasswordHash,
		Picture:      s.Picture,
		LastUpdate:   timestamppb.New(s.LastUpdate),
	}
}
