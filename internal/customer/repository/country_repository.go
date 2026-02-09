package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tokyoyuan/dvd-rental/internal/customer/model"
	"github.com/tokyoyuan/dvd-rental/internal/customer/repository/sqlcgen"
)

// CountryRepository defines read-only data-access operations for countries.
type CountryRepository interface {
	GetCountry(ctx context.Context, countryID int32) (model.Country, error)
	ListCountries(ctx context.Context, limit, offset int32) ([]model.Country, error)
	CountCountries(ctx context.Context) (int64, error)
}

type countryRepository struct {
	q *sqlcgen.Queries
}

// NewCountryRepository creates a new CountryRepository.
func NewCountryRepository(pool *pgxpool.Pool) CountryRepository {
	return &countryRepository{q: sqlcgen.New(pool)}
}

func (r *countryRepository) GetCountry(ctx context.Context, countryID int32) (model.Country, error) {
	row, err := r.q.GetCountry(ctx, countryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Country{}, ErrNotFound
		}
		return model.Country{}, fmt.Errorf("get country: %w", err)
	}
	return toCountryModel(row), nil
}

func (r *countryRepository) ListCountries(ctx context.Context, limit, offset int32) ([]model.Country, error) {
	rows, err := r.q.ListCountries(ctx, sqlcgen.ListCountriesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list countries: %w", err)
	}
	countries := make([]model.Country, len(rows))
	for i, row := range rows {
		countries[i] = toCountryModel(row)
	}
	return countries, nil
}

func (r *countryRepository) CountCountries(ctx context.Context) (int64, error) {
	count, err := r.q.CountCountries(ctx)
	if err != nil {
		return 0, fmt.Errorf("count countries: %w", err)
	}
	return count, nil
}

func toCountryModel(c sqlcgen.Country) model.Country {
	return model.Country{
		CountryID:  c.CountryID,
		Country:    c.Country,
		LastUpdate: timestamptzToTime(c.LastUpdate),
	}
}
