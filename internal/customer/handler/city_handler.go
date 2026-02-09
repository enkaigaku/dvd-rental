package handler

import (
	"context"

	customerv1 "github.com/tokyoyuan/dvd-rental/gen/proto/customer/v1"
	"github.com/tokyoyuan/dvd-rental/internal/customer/service"
)

// CityHandler implements the CityService gRPC server.
type CityHandler struct {
	customerv1.UnimplementedCityServiceServer
	svc *service.CityService
}

// NewCityHandler creates a new CityHandler.
func NewCityHandler(svc *service.CityService) *CityHandler {
	return &CityHandler{svc: svc}
}

func (h *CityHandler) GetCity(ctx context.Context, req *customerv1.GetCityRequest) (*customerv1.City, error) {
	city, err := h.svc.GetCity(ctx, req.GetCityId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return cityToProto(city), nil
}

func (h *CityHandler) ListCities(ctx context.Context, req *customerv1.ListCitiesRequest) (*customerv1.ListCitiesResponse, error) {
	cities, total, err := h.svc.ListCities(ctx, req.GetPageSize(), req.GetPage())
	if err != nil {
		return nil, toGRPCError(err)
	}
	protos := make([]*customerv1.City, len(cities))
	for i, c := range cities {
		protos[i] = cityToProto(c)
	}
	return &customerv1.ListCitiesResponse{
		Cities:     protos,
		TotalCount: int32(total),
	}, nil
}
