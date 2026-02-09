package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/tokyoyuan/dvd-rental/internal/rental/model"
	"github.com/tokyoyuan/dvd-rental/internal/rental/repository"
)

// RentalService contains business logic for rental operations.
type RentalService struct {
	rentalRepo    repository.RentalRepository
	inventoryRepo repository.InventoryRepository
}

// NewRentalService creates a new RentalService.
func NewRentalService(
	rentalRepo repository.RentalRepository,
	inventoryRepo repository.InventoryRepository,
) *RentalService {
	return &RentalService{
		rentalRepo:    rentalRepo,
		inventoryRepo: inventoryRepo,
	}
}

// GetRental returns a rental with enriched details (customer name, film title, store).
func (s *RentalService) GetRental(ctx context.Context, rentalID int32) (model.RentalDetail, error) {
	if rentalID <= 0 {
		return model.RentalDetail{}, fmt.Errorf("rental_id must be positive: %w", ErrInvalidArgument)
	}

	rental, err := s.rentalRepo.GetRental(ctx, rentalID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.RentalDetail{}, fmt.Errorf("rental %d: %w", rentalID, ErrNotFound)
		}
		return model.RentalDetail{}, err
	}

	detail := model.RentalDetail{Rental: rental}

	// Aggregate customer name.
	name, err := s.rentalRepo.GetCustomerName(ctx, rental.CustomerID)
	if err == nil {
		detail.CustomerName = name
	}

	// Aggregate film title and store_id via inventory.
	title, storeID, err := s.rentalRepo.GetFilmTitleByInventory(ctx, rental.InventoryID)
	if err == nil {
		detail.FilmTitle = title
		detail.StoreID = storeID
	}

	return detail, nil
}

// ListRentals returns a paginated list of rentals.
func (s *RentalService) ListRentals(ctx context.Context, pageSize, page int32) ([]model.Rental, int64, error) {
	pageSize, page = clampPagination(pageSize, page)
	offset := (page - 1) * pageSize

	rentals, err := s.rentalRepo.ListRentals(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.rentalRepo.CountRentals(ctx)
	if err != nil {
		return nil, 0, err
	}

	return rentals, total, nil
}

// ListRentalsByCustomer returns rentals for a given customer.
func (s *RentalService) ListRentalsByCustomer(ctx context.Context, customerID, pageSize, page int32) ([]model.Rental, int64, error) {
	if customerID <= 0 {
		return nil, 0, fmt.Errorf("customer_id must be positive: %w", ErrInvalidArgument)
	}

	pageSize, page = clampPagination(pageSize, page)
	offset := (page - 1) * pageSize

	rentals, err := s.rentalRepo.ListRentalsByCustomer(ctx, customerID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.rentalRepo.CountRentalsByCustomer(ctx, customerID)
	if err != nil {
		return nil, 0, err
	}

	return rentals, total, nil
}

// ListRentalsByInventory returns rentals for a given inventory item.
func (s *RentalService) ListRentalsByInventory(ctx context.Context, inventoryID, pageSize, page int32) ([]model.Rental, int64, error) {
	if inventoryID <= 0 {
		return nil, 0, fmt.Errorf("inventory_id must be positive: %w", ErrInvalidArgument)
	}

	pageSize, page = clampPagination(pageSize, page)
	offset := (page - 1) * pageSize

	rentals, err := s.rentalRepo.ListRentalsByInventory(ctx, inventoryID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.rentalRepo.CountRentalsByInventory(ctx, inventoryID)
	if err != nil {
		return nil, 0, err
	}

	return rentals, total, nil
}

// ListOverdueRentals returns rentals that have not been returned.
func (s *RentalService) ListOverdueRentals(ctx context.Context, pageSize, page int32) ([]model.Rental, int64, error) {
	pageSize, page = clampPagination(pageSize, page)
	offset := (page - 1) * pageSize

	rentals, err := s.rentalRepo.ListOverdueRentals(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.rentalRepo.CountOverdueRentals(ctx)
	if err != nil {
		return nil, 0, err
	}

	return rentals, total, nil
}

// CreateRental creates a new rental after validating inventory availability.
func (s *RentalService) CreateRental(ctx context.Context, params repository.CreateRentalParams) (model.Rental, error) {
	if params.InventoryID <= 0 {
		return model.Rental{}, fmt.Errorf("inventory_id must be positive: %w", ErrInvalidArgument)
	}
	if params.CustomerID <= 0 {
		return model.Rental{}, fmt.Errorf("customer_id must be positive: %w", ErrInvalidArgument)
	}
	if params.StaffID <= 0 {
		return model.Rental{}, fmt.Errorf("staff_id must be positive: %w", ErrInvalidArgument)
	}

	// Verify the inventory item exists.
	_, err := s.inventoryRepo.GetInventory(ctx, params.InventoryID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Rental{}, fmt.Errorf("inventory %d not found: %w", params.InventoryID, ErrInvalidArgument)
		}
		return model.Rental{}, err
	}

	// Check availability.
	available, err := s.rentalRepo.IsInventoryAvailable(ctx, params.InventoryID)
	if err != nil {
		return model.Rental{}, err
	}
	if !available {
		return model.Rental{}, fmt.Errorf("inventory %d is currently rented out: %w", params.InventoryID, ErrInvalidArgument)
	}

	rental, err := s.rentalRepo.CreateRental(ctx, params)
	if err != nil {
		if isForeignKeyViolation(err) {
			return model.Rental{}, fmt.Errorf("invalid customer_id or staff_id: %w", ErrInvalidArgument)
		}
		if isUniqueViolation(err) {
			return model.Rental{}, fmt.Errorf("rental already exists: %w", ErrAlreadyExists)
		}
		return model.Rental{}, err
	}
	return rental, nil
}

// ReturnRental marks a rental as returned. Fails if not found or already returned.
func (s *RentalService) ReturnRental(ctx context.Context, rentalID int32) (model.Rental, error) {
	if rentalID <= 0 {
		return model.Rental{}, fmt.Errorf("rental_id must be positive: %w", ErrInvalidArgument)
	}

	rental, err := s.rentalRepo.ReturnRental(ctx, rentalID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Rental{}, fmt.Errorf("rental %d not found or already returned: %w", rentalID, ErrNotFound)
		}
		return model.Rental{}, err
	}
	return rental, nil
}

// DeleteRental deletes a rental. Fails if payment records reference it.
func (s *RentalService) DeleteRental(ctx context.Context, rentalID int32) error {
	if rentalID <= 0 {
		return fmt.Errorf("rental_id must be positive: %w", ErrInvalidArgument)
	}

	// Verify the rental exists.
	_, err := s.rentalRepo.GetRental(ctx, rentalID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("rental %d: %w", rentalID, ErrNotFound)
		}
		return err
	}

	if err := s.rentalRepo.DeleteRental(ctx, rentalID); err != nil {
		if isForeignKeyViolation(err) {
			return fmt.Errorf("rental %d is referenced by payment records: %w", rentalID, ErrForeignKey)
		}
		return err
	}
	return nil
}
