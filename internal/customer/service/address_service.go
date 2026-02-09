package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/tokyoyuan/dvd-rental/internal/customer/model"
	"github.com/tokyoyuan/dvd-rental/internal/customer/repository"
)

// AddressService contains business logic for address operations.
type AddressService struct {
	addressRepo repository.AddressRepository
	cityRepo    repository.CityRepository
}

// NewAddressService creates a new AddressService.
func NewAddressService(addressRepo repository.AddressRepository, cityRepo repository.CityRepository) *AddressService {
	return &AddressService{
		addressRepo: addressRepo,
		cityRepo:    cityRepo,
	}
}

// GetAddress returns an address by ID.
func (s *AddressService) GetAddress(ctx context.Context, addressID int32) (model.Address, error) {
	if addressID <= 0 {
		return model.Address{}, fmt.Errorf("address_id must be positive: %w", ErrInvalidArgument)
	}

	addr, err := s.addressRepo.GetAddress(ctx, addressID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Address{}, fmt.Errorf("address %d: %w", addressID, ErrNotFound)
		}
		return model.Address{}, err
	}
	return addr, nil
}

// ListAddresses returns a paginated list of addresses.
func (s *AddressService) ListAddresses(ctx context.Context, pageSize, page int32) ([]model.Address, int64, error) {
	pageSize, page = clampPagination(pageSize, page)
	offset := (page - 1) * pageSize

	addresses, err := s.addressRepo.ListAddresses(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.addressRepo.CountAddresses(ctx)
	if err != nil {
		return nil, 0, err
	}

	return addresses, total, nil
}

// CreateAddress creates a new address after validation.
func (s *AddressService) CreateAddress(ctx context.Context, params repository.CreateAddressParams) (model.Address, error) {
	if err := s.validateAddressParams(ctx, params.Address, params.District, params.Phone, params.CityID); err != nil {
		return model.Address{}, err
	}

	addr, err := s.addressRepo.CreateAddress(ctx, params)
	if err != nil {
		if isForeignKeyViolation(err) {
			return model.Address{}, fmt.Errorf("invalid city_id: %w", ErrInvalidArgument)
		}
		return model.Address{}, err
	}
	return addr, nil
}

// UpdateAddress updates an existing address after validation.
func (s *AddressService) UpdateAddress(ctx context.Context, params repository.UpdateAddressParams) (model.Address, error) {
	if params.AddressID <= 0 {
		return model.Address{}, fmt.Errorf("address_id must be positive: %w", ErrInvalidArgument)
	}

	if err := s.validateAddressParams(ctx, params.Address, params.District, params.Phone, params.CityID); err != nil {
		return model.Address{}, err
	}

	addr, err := s.addressRepo.UpdateAddress(ctx, params)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Address{}, fmt.Errorf("address %d: %w", params.AddressID, ErrNotFound)
		}
		if isForeignKeyViolation(err) {
			return model.Address{}, fmt.Errorf("invalid city_id: %w", ErrInvalidArgument)
		}
		return model.Address{}, err
	}
	return addr, nil
}

// DeleteAddress deletes an address. Fails if customers reference it.
func (s *AddressService) DeleteAddress(ctx context.Context, addressID int32) error {
	if addressID <= 0 {
		return fmt.Errorf("address_id must be positive: %w", ErrInvalidArgument)
	}

	_, err := s.addressRepo.GetAddress(ctx, addressID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("address %d: %w", addressID, ErrNotFound)
		}
		return err
	}

	if err := s.addressRepo.DeleteAddress(ctx, addressID); err != nil {
		if isForeignKeyViolation(err) {
			return fmt.Errorf("address %d is referenced by customer or store records: %w", addressID, ErrForeignKey)
		}
		return err
	}
	return nil
}

func (s *AddressService) validateAddressParams(ctx context.Context, address, district, phone string, cityID int32) error {
	if address == "" {
		return fmt.Errorf("address must not be empty: %w", ErrInvalidArgument)
	}
	if district == "" {
		return fmt.Errorf("district must not be empty: %w", ErrInvalidArgument)
	}
	if phone == "" {
		return fmt.Errorf("phone must not be empty: %w", ErrInvalidArgument)
	}
	if cityID <= 0 {
		return fmt.Errorf("city_id must be positive: %w", ErrInvalidArgument)
	}

	// Verify city exists.
	if _, err := s.cityRepo.GetCity(ctx, cityID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("city %d not found: %w", cityID, ErrInvalidArgument)
		}
		return err
	}

	return nil
}
