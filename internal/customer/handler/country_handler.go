package handler

import (
	"context"

	customerv1 "github.com/tokyoyuan/dvd-rental/gen/proto/customer/v1"
	"github.com/tokyoyuan/dvd-rental/internal/customer/service"
)

// CountryHandler implements the CountryService gRPC server.
type CountryHandler struct {
	customerv1.UnimplementedCountryServiceServer
	svc *service.CountryService
}

// NewCountryHandler creates a new CountryHandler.
func NewCountryHandler(svc *service.CountryService) *CountryHandler {
	return &CountryHandler{svc: svc}
}

func (h *CountryHandler) GetCountry(ctx context.Context, req *customerv1.GetCountryRequest) (*customerv1.Country, error) {
	country, err := h.svc.GetCountry(ctx, req.GetCountryId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return countryToProto(country), nil
}

func (h *CountryHandler) ListCountries(ctx context.Context, req *customerv1.ListCountriesRequest) (*customerv1.ListCountriesResponse, error) {
	countries, total, err := h.svc.ListCountries(ctx, req.GetPageSize(), req.GetPage())
	if err != nil {
		return nil, toGRPCError(err)
	}
	protos := make([]*customerv1.Country, len(countries))
	for i, c := range countries {
		protos[i] = countryToProto(c)
	}
	return &customerv1.ListCountriesResponse{
		Countries:  protos,
		TotalCount: int32(total),
	}, nil
}
