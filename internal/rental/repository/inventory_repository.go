package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tokyoyuan/dvd-rental/internal/rental/model"
	"github.com/tokyoyuan/dvd-rental/internal/rental/repository/sqlcgen"
)

// CreateInventoryParams holds parameters for creating an inventory item.
type CreateInventoryParams struct {
	FilmID  int32
	StoreID int32
}

// InventoryRepository defines data-access operations for inventory.
type InventoryRepository interface {
	GetInventory(ctx context.Context, inventoryID int32) (model.Inventory, error)
	ListInventory(ctx context.Context, limit, offset int32) ([]model.Inventory, error)
	CountInventory(ctx context.Context) (int64, error)
	ListInventoryByFilm(ctx context.Context, filmID, limit, offset int32) ([]model.Inventory, error)
	CountInventoryByFilm(ctx context.Context, filmID int32) (int64, error)
	ListInventoryByStore(ctx context.Context, storeID, limit, offset int32) ([]model.Inventory, error)
	CountInventoryByStore(ctx context.Context, storeID int32) (int64, error)
	ListAvailableInventory(ctx context.Context, filmID, storeID, limit, offset int32) ([]model.Inventory, error)
	CountAvailableInventory(ctx context.Context, filmID, storeID int32) (int64, error)
	CreateInventory(ctx context.Context, params CreateInventoryParams) (model.Inventory, error)
	DeleteInventory(ctx context.Context, inventoryID int32) error
}

type inventoryRepository struct {
	q *sqlcgen.Queries
}

// NewInventoryRepository creates a new InventoryRepository.
func NewInventoryRepository(pool *pgxpool.Pool) InventoryRepository {
	return &inventoryRepository{q: sqlcgen.New(pool)}
}

func (r *inventoryRepository) GetInventory(ctx context.Context, inventoryID int32) (model.Inventory, error) {
	row, err := r.q.GetInventory(ctx, inventoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Inventory{}, ErrNotFound
		}
		return model.Inventory{}, fmt.Errorf("get inventory: %w", err)
	}
	return toInventoryModel(row), nil
}

func (r *inventoryRepository) ListInventory(ctx context.Context, limit, offset int32) ([]model.Inventory, error) {
	rows, err := r.q.ListInventory(ctx, sqlcgen.ListInventoryParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("list inventory: %w", err)
	}
	return toInventoryModels(rows), nil
}

func (r *inventoryRepository) CountInventory(ctx context.Context) (int64, error) {
	count, err := r.q.CountInventory(ctx)
	if err != nil {
		return 0, fmt.Errorf("count inventory: %w", err)
	}
	return count, nil
}

func (r *inventoryRepository) ListInventoryByFilm(ctx context.Context, filmID, limit, offset int32) ([]model.Inventory, error) {
	rows, err := r.q.ListInventoryByFilm(ctx, sqlcgen.ListInventoryByFilmParams{
		FilmID: filmID, Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list inventory by film: %w", err)
	}
	return toInventoryModels(rows), nil
}

func (r *inventoryRepository) CountInventoryByFilm(ctx context.Context, filmID int32) (int64, error) {
	count, err := r.q.CountInventoryByFilm(ctx, filmID)
	if err != nil {
		return 0, fmt.Errorf("count inventory by film: %w", err)
	}
	return count, nil
}

func (r *inventoryRepository) ListInventoryByStore(ctx context.Context, storeID, limit, offset int32) ([]model.Inventory, error) {
	rows, err := r.q.ListInventoryByStore(ctx, sqlcgen.ListInventoryByStoreParams{
		StoreID: storeID, Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list inventory by store: %w", err)
	}
	return toInventoryModels(rows), nil
}

func (r *inventoryRepository) CountInventoryByStore(ctx context.Context, storeID int32) (int64, error) {
	count, err := r.q.CountInventoryByStore(ctx, storeID)
	if err != nil {
		return 0, fmt.Errorf("count inventory by store: %w", err)
	}
	return count, nil
}

func (r *inventoryRepository) ListAvailableInventory(ctx context.Context, filmID, storeID, limit, offset int32) ([]model.Inventory, error) {
	rows, err := r.q.ListAvailableInventory(ctx, sqlcgen.ListAvailableInventoryParams{
		FilmID: filmID, StoreID: storeID, Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list available inventory: %w", err)
	}
	return toInventoryModels(rows), nil
}

func (r *inventoryRepository) CountAvailableInventory(ctx context.Context, filmID, storeID int32) (int64, error) {
	count, err := r.q.CountAvailableInventory(ctx, sqlcgen.CountAvailableInventoryParams{
		FilmID: filmID, StoreID: storeID,
	})
	if err != nil {
		return 0, fmt.Errorf("count available inventory: %w", err)
	}
	return count, nil
}

func (r *inventoryRepository) CreateInventory(ctx context.Context, params CreateInventoryParams) (model.Inventory, error) {
	row, err := r.q.CreateInventory(ctx, sqlcgen.CreateInventoryParams{
		FilmID:  params.FilmID,
		StoreID: params.StoreID,
	})
	if err != nil {
		return model.Inventory{}, fmt.Errorf("create inventory: %w", err)
	}
	return toInventoryModel(row), nil
}

func (r *inventoryRepository) DeleteInventory(ctx context.Context, inventoryID int32) error {
	if err := r.q.DeleteInventory(ctx, inventoryID); err != nil {
		return fmt.Errorf("delete inventory: %w", err)
	}
	return nil
}

func toInventoryModel(i sqlcgen.Inventory) model.Inventory {
	return model.Inventory{
		InventoryID: i.InventoryID,
		FilmID:      i.FilmID,
		StoreID:     i.StoreID,
		LastUpdate:  timestamptzToTime(i.LastUpdate),
	}
}

func toInventoryModels(rows []sqlcgen.Inventory) []model.Inventory {
	items := make([]model.Inventory, len(rows))
	for i, row := range rows {
		items[i] = toInventoryModel(row)
	}
	return items
}
