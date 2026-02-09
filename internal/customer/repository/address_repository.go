package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/enkaigaku/dvd-rental/internal/customer/model"
	"github.com/enkaigaku/dvd-rental/internal/customer/repository/sqlcgen"
)

// CreateAddressParams holds parameters for creating an address.
type CreateAddressParams struct {
	Address    string
	Address2   string
	District   string
	CityID     int32
	PostalCode string
	Phone      string
}

// UpdateAddressParams holds parameters for updating an address.
type UpdateAddressParams struct {
	AddressID  int32
	Address    string
	Address2   string
	District   string
	CityID     int32
	PostalCode string
	Phone      string
}

// AddressRepository defines data-access operations for addresses.
type AddressRepository interface {
	GetAddress(ctx context.Context, addressID int32) (model.Address, error)
	ListAddresses(ctx context.Context, limit, offset int32) ([]model.Address, error)
	CountAddresses(ctx context.Context) (int64, error)
	CreateAddress(ctx context.Context, params CreateAddressParams) (model.Address, error)
	UpdateAddress(ctx context.Context, params UpdateAddressParams) (model.Address, error)
	DeleteAddress(ctx context.Context, addressID int32) error
}

type addressRepository struct {
	q *sqlcgen.Queries
}

// NewAddressRepository creates a new AddressRepository.
func NewAddressRepository(pool *pgxpool.Pool) AddressRepository {
	return &addressRepository{q: sqlcgen.New(pool)}
}

func (r *addressRepository) GetAddress(ctx context.Context, addressID int32) (model.Address, error) {
	row, err := r.q.GetAddress(ctx, addressID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Address{}, ErrNotFound
		}
		return model.Address{}, fmt.Errorf("get address: %w", err)
	}
	return toAddressModel(row), nil
}

func (r *addressRepository) ListAddresses(ctx context.Context, limit, offset int32) ([]model.Address, error) {
	rows, err := r.q.ListAddresses(ctx, sqlcgen.ListAddressesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list addresses: %w", err)
	}
	addresses := make([]model.Address, len(rows))
	for i, row := range rows {
		addresses[i] = toAddressModel(row)
	}
	return addresses, nil
}

func (r *addressRepository) CountAddresses(ctx context.Context) (int64, error) {
	count, err := r.q.CountAddresses(ctx)
	if err != nil {
		return 0, fmt.Errorf("count addresses: %w", err)
	}
	return count, nil
}

func (r *addressRepository) CreateAddress(ctx context.Context, params CreateAddressParams) (model.Address, error) {
	row, err := r.q.CreateAddress(ctx, sqlcgen.CreateAddressParams{
		Address:    params.Address,
		Address2:   stringToText(params.Address2),
		District:   params.District,
		CityID:     params.CityID,
		PostalCode: stringToText(params.PostalCode),
		Phone:      params.Phone,
	})
	if err != nil {
		return model.Address{}, fmt.Errorf("create address: %w", err)
	}
	return toAddressModel(row), nil
}

func (r *addressRepository) UpdateAddress(ctx context.Context, params UpdateAddressParams) (model.Address, error) {
	row, err := r.q.UpdateAddress(ctx, sqlcgen.UpdateAddressParams{
		AddressID:  params.AddressID,
		Address:    params.Address,
		Address2:   stringToText(params.Address2),
		District:   params.District,
		CityID:     params.CityID,
		PostalCode: stringToText(params.PostalCode),
		Phone:      params.Phone,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Address{}, ErrNotFound
		}
		return model.Address{}, fmt.Errorf("update address: %w", err)
	}
	return toAddressModel(row), nil
}

func (r *addressRepository) DeleteAddress(ctx context.Context, addressID int32) error {
	if err := r.q.DeleteAddress(ctx, addressID); err != nil {
		return fmt.Errorf("delete address: %w", err)
	}
	return nil
}

func toAddressModel(a sqlcgen.Address) model.Address {
	return model.Address{
		AddressID:  a.AddressID,
		Address:    a.Address,
		Address2:   textToString(a.Address2),
		District:   a.District,
		CityID:     a.CityID,
		PostalCode: textToString(a.PostalCode),
		Phone:      a.Phone,
		LastUpdate: timestamptzToTime(a.LastUpdate),
	}
}
