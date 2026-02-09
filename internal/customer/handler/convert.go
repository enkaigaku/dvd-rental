package handler

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	customerv1 "github.com/enkaigaku/dvd-rental/gen/proto/customer/v1"
	"github.com/enkaigaku/dvd-rental/internal/customer/model"
	"github.com/enkaigaku/dvd-rental/internal/customer/service"
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

func customerToProto(c model.Customer) *customerv1.Customer {
	return &customerv1.Customer{
		CustomerId:   c.CustomerID,
		StoreId:      c.StoreID,
		FirstName:    c.FirstName,
		LastName:     c.LastName,
		Email:        c.Email,
		AddressId:    c.AddressID,
		Active:       c.Active,
		CreateDate:   timestamppb.New(c.CreateDate),
		LastUpdate:   timestamppb.New(c.LastUpdate),
		PasswordHash: c.PasswordHash,
	}
}

func customerDetailToProto(d model.CustomerDetail) *customerv1.CustomerDetail {
	return &customerv1.CustomerDetail{
		Customer:   customerToProto(d.Customer),
		Address:    d.Address,
		Address2:   d.Address2,
		District:   d.District,
		City:       d.CityName,
		Country:    d.CountryName,
		PostalCode: d.PostalCode,
		Phone:      d.Phone,
	}
}

func addressToProto(a model.Address) *customerv1.Address {
	return &customerv1.Address{
		AddressId:  a.AddressID,
		Address:    a.Address,
		Address2:   a.Address2,
		District:   a.District,
		CityId:     a.CityID,
		PostalCode: a.PostalCode,
		Phone:      a.Phone,
		LastUpdate: timestamppb.New(a.LastUpdate),
	}
}

func cityToProto(c model.City) *customerv1.City {
	return &customerv1.City{
		CityId:     c.CityID,
		City:       c.City,
		CountryId:  c.CountryID,
		LastUpdate: timestamppb.New(c.LastUpdate),
	}
}

func countryToProto(c model.Country) *customerv1.Country {
	return &customerv1.Country{
		CountryId:  c.CountryID,
		Country:    c.Country,
		LastUpdate: timestamppb.New(c.LastUpdate),
	}
}
