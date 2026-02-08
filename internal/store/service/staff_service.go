package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/tokyoyuan/dvd-rental/internal/store/model"
	"github.com/tokyoyuan/dvd-rental/internal/store/repository"
)

// StaffService handles business logic for staff operations.
type StaffService struct {
	staffRepo repository.StaffRepository
	storeRepo repository.StoreRepository
}

// NewStaffService creates a new StaffService.
func NewStaffService(staffRepo repository.StaffRepository, storeRepo repository.StoreRepository) *StaffService {
	return &StaffService{
		staffRepo: staffRepo,
		storeRepo: storeRepo,
	}
}

// GetStaff retrieves a staff member by ID (includes picture, excludes password_hash).
func (s *StaffService) GetStaff(ctx context.Context, staffID int32) (model.Staff, error) {
	if staffID <= 0 {
		return model.Staff{}, fmt.Errorf("staff_id must be positive: %w", ErrInvalidArgument)
	}
	staff, err := s.staffRepo.GetStaff(ctx, staffID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Staff{}, fmt.Errorf("staff %d: %w", staffID, ErrNotFound)
		}
		return model.Staff{}, err
	}
	return staff, nil
}

// GetStaffByUsername retrieves a staff member by username (includes password_hash for auth, excludes picture).
func (s *StaffService) GetStaffByUsername(ctx context.Context, username string) (model.Staff, error) {
	if username == "" {
		return model.Staff{}, fmt.Errorf("username must not be empty: %w", ErrInvalidArgument)
	}
	staff, err := s.staffRepo.GetStaffByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Staff{}, fmt.Errorf("staff with username %q: %w", username, ErrNotFound)
		}
		return model.Staff{}, err
	}
	return staff, nil
}

// ListStaff retrieves a paginated list of staff, optionally filtered by active status.
func (s *StaffService) ListStaff(ctx context.Context, pageSize, page int32, activeOnly bool) ([]model.Staff, int64, error) {
	pageSize, page = clampPagination(pageSize, page)
	offset := (page - 1) * pageSize

	var (
		staff []model.Staff
		total int64
		err   error
	)

	if activeOnly {
		staff, err = s.staffRepo.ListActiveStaff(ctx, pageSize, offset)
		if err != nil {
			return nil, 0, err
		}
		total, err = s.staffRepo.CountActiveStaff(ctx)
	} else {
		staff, err = s.staffRepo.ListStaff(ctx, pageSize, offset)
		if err != nil {
			return nil, 0, err
		}
		total, err = s.staffRepo.CountStaff(ctx)
	}
	if err != nil {
		return nil, 0, err
	}
	return staff, total, nil
}

// ListStaffByStore retrieves a paginated list of staff for a specific store.
func (s *StaffService) ListStaffByStore(ctx context.Context, storeID, pageSize, page int32, activeOnly bool) ([]model.Staff, int64, error) {
	if storeID <= 0 {
		return nil, 0, fmt.Errorf("store_id must be positive: %w", ErrInvalidArgument)
	}
	pageSize, page = clampPagination(pageSize, page)
	offset := (page - 1) * pageSize

	var (
		staff []model.Staff
		total int64
		err   error
	)

	if activeOnly {
		staff, err = s.staffRepo.ListActiveStaffByStore(ctx, storeID, pageSize, offset)
		if err != nil {
			return nil, 0, err
		}
		total, err = s.staffRepo.CountActiveStaffByStore(ctx, storeID)
	} else {
		staff, err = s.staffRepo.ListStaffByStore(ctx, storeID, pageSize, offset)
		if err != nil {
			return nil, 0, err
		}
		total, err = s.staffRepo.CountStaffByStore(ctx, storeID)
	}
	if err != nil {
		return nil, 0, err
	}
	return staff, total, nil
}

// CreateStaff creates a new staff member after validating the store exists.
func (s *StaffService) CreateStaff(ctx context.Context, params repository.CreateStaffParams) (model.Staff, error) {
	if params.FirstName == "" {
		return model.Staff{}, fmt.Errorf("first_name must not be empty: %w", ErrInvalidArgument)
	}
	if params.LastName == "" {
		return model.Staff{}, fmt.Errorf("last_name must not be empty: %w", ErrInvalidArgument)
	}
	if params.AddressID <= 0 {
		return model.Staff{}, fmt.Errorf("address_id must be positive: %w", ErrInvalidArgument)
	}
	if params.StoreID <= 0 {
		return model.Staff{}, fmt.Errorf("store_id must be positive: %w", ErrInvalidArgument)
	}
	if params.Username == "" {
		return model.Staff{}, fmt.Errorf("username must not be empty: %w", ErrInvalidArgument)
	}

	// Validate store exists.
	_, err := s.storeRepo.GetStore(ctx, params.StoreID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Staff{}, fmt.Errorf("store %d not found: %w", params.StoreID, ErrInvalidArgument)
		}
		return model.Staff{}, err
	}

	staff, err := s.staffRepo.CreateStaff(ctx, params)
	if err != nil {
		if isUniqueViolation(err) {
			return model.Staff{}, fmt.Errorf("username %q already taken: %w", params.Username, ErrAlreadyExists)
		}
		return model.Staff{}, err
	}
	return staff, nil
}

// UpdateStaff updates a staff member's information.
func (s *StaffService) UpdateStaff(ctx context.Context, params repository.UpdateStaffParams) (model.Staff, error) {
	if params.StaffID <= 0 {
		return model.Staff{}, fmt.Errorf("staff_id must be positive: %w", ErrInvalidArgument)
	}
	if params.FirstName == "" {
		return model.Staff{}, fmt.Errorf("first_name must not be empty: %w", ErrInvalidArgument)
	}
	if params.LastName == "" {
		return model.Staff{}, fmt.Errorf("last_name must not be empty: %w", ErrInvalidArgument)
	}
	if params.AddressID <= 0 {
		return model.Staff{}, fmt.Errorf("address_id must be positive: %w", ErrInvalidArgument)
	}
	if params.StoreID <= 0 {
		return model.Staff{}, fmt.Errorf("store_id must be positive: %w", ErrInvalidArgument)
	}
	if params.Username == "" {
		return model.Staff{}, fmt.Errorf("username must not be empty: %w", ErrInvalidArgument)
	}

	// Validate store exists.
	_, err := s.storeRepo.GetStore(ctx, params.StoreID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Staff{}, fmt.Errorf("store %d not found: %w", params.StoreID, ErrInvalidArgument)
		}
		return model.Staff{}, err
	}

	staff, err := s.staffRepo.UpdateStaff(ctx, params)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Staff{}, fmt.Errorf("staff %d: %w", params.StaffID, ErrNotFound)
		}
		if isUniqueViolation(err) {
			return model.Staff{}, fmt.Errorf("username %q already taken: %w", params.Username, ErrAlreadyExists)
		}
		return model.Staff{}, err
	}
	return staff, nil
}

// DeactivateStaff deactivates a staff member (soft delete).
func (s *StaffService) DeactivateStaff(ctx context.Context, staffID int32) (model.Staff, error) {
	if staffID <= 0 {
		return model.Staff{}, fmt.Errorf("staff_id must be positive: %w", ErrInvalidArgument)
	}
	staff, err := s.staffRepo.DeactivateStaff(ctx, staffID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Staff{}, fmt.Errorf("staff %d: %w", staffID, ErrNotFound)
		}
		return model.Staff{}, err
	}
	return staff, nil
}

// UpdateStaffPassword updates a staff member's password hash.
func (s *StaffService) UpdateStaffPassword(ctx context.Context, staffID int32, passwordHash string) error {
	if staffID <= 0 {
		return fmt.Errorf("staff_id must be positive: %w", ErrInvalidArgument)
	}
	if passwordHash == "" {
		return fmt.Errorf("password_hash must not be empty: %w", ErrInvalidArgument)
	}

	// Verify staff exists.
	_, err := s.staffRepo.GetStaff(ctx, staffID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("staff %d: %w", staffID, ErrNotFound)
		}
		return err
	}

	return s.staffRepo.UpdateStaffPassword(ctx, staffID, passwordHash)
}
