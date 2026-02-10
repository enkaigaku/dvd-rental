// Package repository provides data access for the store service.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/enkaigaku/dvd-rental/internal/store/model"
	"github.com/enkaigaku/dvd-rental/gen/sqlc/store"
)

// ErrNotFound is returned when a requested record does not exist.
var ErrNotFound = errors.New("record not found")

// StoreRepository defines the data access interface for stores.
type StoreRepository interface {
	GetStore(ctx context.Context, storeID int32) (model.Store, error)
	ListStores(ctx context.Context, limit, offset int32) ([]model.Store, error)
	CountStores(ctx context.Context) (int64, error)
	CreateStore(ctx context.Context, managerStaffID, addressID int32) (model.Store, error)
	UpdateStore(ctx context.Context, storeID, managerStaffID, addressID int32) (model.Store, error)
	DeleteStore(ctx context.Context, storeID int32) error
}

type storeRepository struct {
	q *storesqlc.Queries
}

// NewStoreRepository creates a new StoreRepository backed by PostgreSQL.
func NewStoreRepository(pool *pgxpool.Pool) StoreRepository {
	return &storeRepository{q: storesqlc.New(pool)}
}

func (r *storeRepository) GetStore(ctx context.Context, storeID int32) (model.Store, error) {
	row, err := r.q.GetStore(ctx, storeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Store{}, ErrNotFound
		}
		return model.Store{}, fmt.Errorf("get store: %w", err)
	}
	return toStoreModel(row), nil
}

func (r *storeRepository) ListStores(ctx context.Context, limit, offset int32) ([]model.Store, error) {
	rows, err := r.q.ListStores(ctx, storesqlc.ListStoresParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list stores: %w", err)
	}
	stores := make([]model.Store, len(rows))
	for i, row := range rows {
		stores[i] = toStoreModel(row)
	}
	return stores, nil
}

func (r *storeRepository) CountStores(ctx context.Context) (int64, error) {
	count, err := r.q.CountStores(ctx)
	if err != nil {
		return 0, fmt.Errorf("count stores: %w", err)
	}
	return count, nil
}

func (r *storeRepository) CreateStore(ctx context.Context, managerStaffID, addressID int32) (model.Store, error) {
	row, err := r.q.CreateStore(ctx, storesqlc.CreateStoreParams{
		ManagerStaffID: managerStaffID,
		AddressID:      addressID,
	})
	if err != nil {
		return model.Store{}, fmt.Errorf("create store: %w", err)
	}
	return toStoreModel(row), nil
}

func (r *storeRepository) UpdateStore(ctx context.Context, storeID, managerStaffID, addressID int32) (model.Store, error) {
	row, err := r.q.UpdateStore(ctx, storesqlc.UpdateStoreParams{
		StoreID:        storeID,
		ManagerStaffID: managerStaffID,
		AddressID:      addressID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Store{}, ErrNotFound
		}
		return model.Store{}, fmt.Errorf("update store: %w", err)
	}
	return toStoreModel(row), nil
}

func (r *storeRepository) DeleteStore(ctx context.Context, storeID int32) error {
	if err := r.q.DeleteStore(ctx, storeID); err != nil {
		return fmt.Errorf("delete store: %w", err)
	}
	return nil
}

func toStoreModel(s storesqlc.Store) model.Store {
	return model.Store{
		StoreID:        s.StoreID,
		ManagerStaffID: s.ManagerStaffID,
		AddressID:      s.AddressID,
		LastUpdate:     s.LastUpdate.Time,
	}
}
