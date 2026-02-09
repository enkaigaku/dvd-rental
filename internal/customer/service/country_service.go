package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/tokyoyuan/dvd-rental/internal/customer/model"
	"github.com/tokyoyuan/dvd-rental/internal/customer/repository"
)

// CountryService contains business logic for country operations (read-only).
type CountryService struct {
	countryRepo repository.CountryRepository
}

// NewCountryService creates a new CountryService.
func NewCountryService(countryRepo repository.CountryRepository) *CountryService {
	return &CountryService{countryRepo: countryRepo}
}

// GetCountry returns a country by ID.
func (s *CountryService) GetCountry(ctx context.Context, countryID int32) (model.Country, error) {
	if countryID <= 0 {
		return model.Country{}, fmt.Errorf("country_id must be positive: %w", ErrInvalidArgument)
	}

	country, err := s.countryRepo.GetCountry(ctx, countryID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Country{}, fmt.Errorf("country %d: %w", countryID, ErrNotFound)
		}
		return model.Country{}, err
	}
	return country, nil
}

// ListCountries returns a paginated list of countries.
func (s *CountryService) ListCountries(ctx context.Context, pageSize, page int32) ([]model.Country, int64, error) {
	pageSize, page = clampPagination(pageSize, page)
	offset := (page - 1) * pageSize

	countries, err := s.countryRepo.ListCountries(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.countryRepo.CountCountries(ctx)
	if err != nil {
		return nil, 0, err
	}

	return countries, total, nil
}
