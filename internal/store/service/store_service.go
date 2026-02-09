package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/enkaigaku/dvd-rental/internal/store/model"
	"github.com/enkaigaku/dvd-rental/internal/store/repository"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

// StoreService handles business logic for store operations.
type StoreService struct {
	storeRepo repository.StoreRepository
	staffRepo repository.StaffRepository
}

// NewStoreService creates a new StoreService.
func NewStoreService(storeRepo repository.StoreRepository, staffRepo repository.StaffRepository) *StoreService {
	return &StoreService{
		storeRepo: storeRepo,
		staffRepo: staffRepo,
	}
}

// GetStore retrieves a store by ID.
func (s *StoreService) GetStore(ctx context.Context, storeID int32) (model.Store, error) {
	if storeID <= 0 {
		return model.Store{}, fmt.Errorf("store_id must be positive: %w", ErrInvalidArgument)
	}
	store, err := s.storeRepo.GetStore(ctx, storeID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Store{}, fmt.Errorf("store %d: %w", storeID, ErrNotFound)
		}
		return model.Store{}, err
	}
	return store, nil
}

// ListStores retrieves a paginated list of stores.
func (s *StoreService) ListStores(ctx context.Context, pageSize, page int32) ([]model.Store, int64, error) {
	pageSize, page = clampPagination(pageSize, page)
	offset := (page - 1) * pageSize

	stores, err := s.storeRepo.ListStores(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.storeRepo.CountStores(ctx)
	if err != nil {
		return nil, 0, err
	}
	return stores, total, nil
}

// CreateStore creates a new store after validating the manager exists and is active.
func (s *StoreService) CreateStore(ctx context.Context, managerStaffID, addressID int32) (model.Store, error) {
	if managerStaffID <= 0 {
		return model.Store{}, fmt.Errorf("manager_staff_id must be positive: %w", ErrInvalidArgument)
	}
	if addressID <= 0 {
		return model.Store{}, fmt.Errorf("address_id must be positive: %w", ErrInvalidArgument)
	}

	// Validate manager staff exists and is active.
	staff, err := s.staffRepo.GetStaff(ctx, managerStaffID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Store{}, fmt.Errorf("manager staff %d not found: %w", managerStaffID, ErrInvalidArgument)
		}
		return model.Store{}, err
	}
	if !staff.Active {
		return model.Store{}, fmt.Errorf("manager staff %d is not active: %w", managerStaffID, ErrInvalidArgument)
	}

	store, err := s.storeRepo.CreateStore(ctx, managerStaffID, addressID)
	if err != nil {
		// Unique constraint violation on manager_staff_id.
		if isUniqueViolation(err) {
			return model.Store{}, fmt.Errorf("staff %d already manages another store: %w", managerStaffID, ErrAlreadyExists)
		}
		return model.Store{}, err
	}
	return store, nil
}

// UpdateStore updates a store after validating the new manager.
func (s *StoreService) UpdateStore(ctx context.Context, storeID, managerStaffID, addressID int32) (model.Store, error) {
	if storeID <= 0 {
		return model.Store{}, fmt.Errorf("store_id must be positive: %w", ErrInvalidArgument)
	}
	if managerStaffID <= 0 {
		return model.Store{}, fmt.Errorf("manager_staff_id must be positive: %w", ErrInvalidArgument)
	}
	if addressID <= 0 {
		return model.Store{}, fmt.Errorf("address_id must be positive: %w", ErrInvalidArgument)
	}

	// Validate manager staff exists and is active.
	staff, err := s.staffRepo.GetStaff(ctx, managerStaffID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Store{}, fmt.Errorf("manager staff %d not found: %w", managerStaffID, ErrInvalidArgument)
		}
		return model.Store{}, err
	}
	if !staff.Active {
		return model.Store{}, fmt.Errorf("manager staff %d is not active: %w", managerStaffID, ErrInvalidArgument)
	}

	store, err := s.storeRepo.UpdateStore(ctx, storeID, managerStaffID, addressID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Store{}, fmt.Errorf("store %d: %w", storeID, ErrNotFound)
		}
		if isUniqueViolation(err) {
			return model.Store{}, fmt.Errorf("staff %d already manages another store: %w", managerStaffID, ErrAlreadyExists)
		}
		return model.Store{}, err
	}
	return store, nil
}

// DeleteStore deletes a store by ID.
func (s *StoreService) DeleteStore(ctx context.Context, storeID int32) error {
	if storeID <= 0 {
		return fmt.Errorf("store_id must be positive: %w", ErrInvalidArgument)
	}

	// Verify the store exists before attempting delete.
	_, err := s.storeRepo.GetStore(ctx, storeID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("store %d: %w", storeID, ErrNotFound)
		}
		return err
	}

	return s.storeRepo.DeleteStore(ctx, storeID)
}

func clampPagination(pageSize, page int32) (int32, int32) {
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	if page <= 0 {
		page = 1
	}
	return pageSize, page
}
