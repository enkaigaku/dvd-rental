package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/enkaigaku/dvd-rental/internal/customer/model"
	"github.com/enkaigaku/dvd-rental/internal/customer/repository"
)

// CityService contains business logic for city operations (read-only).
type CityService struct {
	cityRepo repository.CityRepository
}

// NewCityService creates a new CityService.
func NewCityService(cityRepo repository.CityRepository) *CityService {
	return &CityService{cityRepo: cityRepo}
}

// GetCity returns a city by ID.
func (s *CityService) GetCity(ctx context.Context, cityID int32) (model.City, error) {
	if cityID <= 0 {
		return model.City{}, fmt.Errorf("city_id must be positive: %w", ErrInvalidArgument)
	}

	city, err := s.cityRepo.GetCity(ctx, cityID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.City{}, fmt.Errorf("city %d: %w", cityID, ErrNotFound)
		}
		return model.City{}, err
	}
	return city, nil
}

// ListCities returns a paginated list of cities.
func (s *CityService) ListCities(ctx context.Context, pageSize, page int32) ([]model.City, int64, error) {
	pageSize, page = clampPagination(pageSize, page)
	offset := (page - 1) * pageSize

	cities, err := s.cityRepo.ListCities(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.cityRepo.CountCities(ctx)
	if err != nil {
		return nil, 0, err
	}

	return cities, total, nil
}
