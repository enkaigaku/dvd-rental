package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/tokyoyuan/dvd-rental/internal/rental/model"
	"github.com/tokyoyuan/dvd-rental/internal/rental/repository"
)

// InventoryService contains business logic for inventory operations.
type InventoryService struct {
	inventoryRepo repository.InventoryRepository
	rentalRepo    repository.RentalRepository
}

// NewInventoryService creates a new InventoryService.
func NewInventoryService(
	inventoryRepo repository.InventoryRepository,
	rentalRepo repository.RentalRepository,
) *InventoryService {
	return &InventoryService{
		inventoryRepo: inventoryRepo,
		rentalRepo:    rentalRepo,
	}
}

// GetInventory returns an inventory item by ID.
func (s *InventoryService) GetInventory(ctx context.Context, inventoryID int32) (model.Inventory, error) {
	if inventoryID <= 0 {
		return model.Inventory{}, fmt.Errorf("inventory_id must be positive: %w", ErrInvalidArgument)
	}

	inv, err := s.inventoryRepo.GetInventory(ctx, inventoryID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Inventory{}, fmt.Errorf("inventory %d: %w", inventoryID, ErrNotFound)
		}
		return model.Inventory{}, err
	}
	return inv, nil
}

// ListInventory returns a paginated list of all inventory items.
func (s *InventoryService) ListInventory(ctx context.Context, pageSize, page int32) ([]model.Inventory, int64, error) {
	pageSize, page = clampPagination(pageSize, page)
	offset := (page - 1) * pageSize

	items, err := s.inventoryRepo.ListInventory(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.inventoryRepo.CountInventory(ctx)
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// ListInventoryByFilm returns inventory for a given film.
func (s *InventoryService) ListInventoryByFilm(ctx context.Context, filmID, pageSize, page int32) ([]model.Inventory, int64, error) {
	if filmID <= 0 {
		return nil, 0, fmt.Errorf("film_id must be positive: %w", ErrInvalidArgument)
	}

	pageSize, page = clampPagination(pageSize, page)
	offset := (page - 1) * pageSize

	items, err := s.inventoryRepo.ListInventoryByFilm(ctx, filmID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.inventoryRepo.CountInventoryByFilm(ctx, filmID)
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// ListInventoryByStore returns inventory for a given store.
func (s *InventoryService) ListInventoryByStore(ctx context.Context, storeID, pageSize, page int32) ([]model.Inventory, int64, error) {
	if storeID <= 0 {
		return nil, 0, fmt.Errorf("store_id must be positive: %w", ErrInvalidArgument)
	}

	pageSize, page = clampPagination(pageSize, page)
	offset := (page - 1) * pageSize

	items, err := s.inventoryRepo.ListInventoryByStore(ctx, storeID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.inventoryRepo.CountInventoryByStore(ctx, storeID)
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// CheckInventoryAvailability checks whether an inventory item is currently available.
func (s *InventoryService) CheckInventoryAvailability(ctx context.Context, inventoryID int32) (bool, error) {
	if inventoryID <= 0 {
		return false, fmt.Errorf("inventory_id must be positive: %w", ErrInvalidArgument)
	}

	// Verify the inventory item exists.
	_, err := s.inventoryRepo.GetInventory(ctx, inventoryID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return false, fmt.Errorf("inventory %d: %w", inventoryID, ErrNotFound)
		}
		return false, err
	}

	return s.rentalRepo.IsInventoryAvailable(ctx, inventoryID)
}

// ListAvailableInventory returns inventory items available for a given film and store.
func (s *InventoryService) ListAvailableInventory(ctx context.Context, filmID, storeID, pageSize, page int32) ([]model.Inventory, int64, error) {
	if filmID <= 0 {
		return nil, 0, fmt.Errorf("film_id must be positive: %w", ErrInvalidArgument)
	}
	if storeID <= 0 {
		return nil, 0, fmt.Errorf("store_id must be positive: %w", ErrInvalidArgument)
	}

	pageSize, page = clampPagination(pageSize, page)
	offset := (page - 1) * pageSize

	items, err := s.inventoryRepo.ListAvailableInventory(ctx, filmID, storeID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.inventoryRepo.CountAvailableInventory(ctx, filmID, storeID)
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// CreateInventory creates a new inventory item.
func (s *InventoryService) CreateInventory(ctx context.Context, params repository.CreateInventoryParams) (model.Inventory, error) {
	if params.FilmID <= 0 {
		return model.Inventory{}, fmt.Errorf("film_id must be positive: %w", ErrInvalidArgument)
	}
	if params.StoreID <= 0 {
		return model.Inventory{}, fmt.Errorf("store_id must be positive: %w", ErrInvalidArgument)
	}

	inv, err := s.inventoryRepo.CreateInventory(ctx, params)
	if err != nil {
		if isForeignKeyViolation(err) {
			return model.Inventory{}, fmt.Errorf("invalid film_id or store_id: %w", ErrInvalidArgument)
		}
		return model.Inventory{}, err
	}
	return inv, nil
}

// DeleteInventory deletes an inventory item. Fails if rental records reference it.
func (s *InventoryService) DeleteInventory(ctx context.Context, inventoryID int32) error {
	if inventoryID <= 0 {
		return fmt.Errorf("inventory_id must be positive: %w", ErrInvalidArgument)
	}

	// Verify the inventory item exists.
	_, err := s.inventoryRepo.GetInventory(ctx, inventoryID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("inventory %d: %w", inventoryID, ErrNotFound)
		}
		return err
	}

	if err := s.inventoryRepo.DeleteInventory(ctx, inventoryID); err != nil {
		if isForeignKeyViolation(err) {
			return fmt.Errorf("inventory %d is referenced by rental records: %w", inventoryID, ErrForeignKey)
		}
		return err
	}
	return nil
}
