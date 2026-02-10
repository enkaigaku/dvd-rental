package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/enkaigaku/dvd-rental/internal/customer/model"
	"github.com/enkaigaku/dvd-rental/gen/sqlc/customer"
)

// CityRepository defines read-only data-access operations for cities.
type CityRepository interface {
	GetCity(ctx context.Context, cityID int32) (model.City, error)
	ListCities(ctx context.Context, limit, offset int32) ([]model.City, error)
	CountCities(ctx context.Context) (int64, error)
}

type cityRepository struct {
	q *customersqlc.Queries
}

// NewCityRepository creates a new CityRepository.
func NewCityRepository(pool *pgxpool.Pool) CityRepository {
	return &cityRepository{q: customersqlc.New(pool)}
}

func (r *cityRepository) GetCity(ctx context.Context, cityID int32) (model.City, error) {
	row, err := r.q.GetCity(ctx, cityID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.City{}, ErrNotFound
		}
		return model.City{}, fmt.Errorf("get city: %w", err)
	}
	return toCityModel(row), nil
}

func (r *cityRepository) ListCities(ctx context.Context, limit, offset int32) ([]model.City, error) {
	rows, err := r.q.ListCities(ctx, customersqlc.ListCitiesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list cities: %w", err)
	}
	cities := make([]model.City, len(rows))
	for i, row := range rows {
		cities[i] = toCityModel(row)
	}
	return cities, nil
}

func (r *cityRepository) CountCities(ctx context.Context) (int64, error) {
	count, err := r.q.CountCities(ctx)
	if err != nil {
		return 0, fmt.Errorf("count cities: %w", err)
	}
	return count, nil
}

func toCityModel(c customersqlc.City) model.City {
	return model.City{
		CityID:     c.CityID,
		City:       c.City,
		CountryID:  c.CountryID,
		LastUpdate: timestamptzToTime(c.LastUpdate),
	}
}
