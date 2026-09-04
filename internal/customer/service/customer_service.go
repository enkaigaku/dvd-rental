package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/enkaigaku/dvd-rental/internal/customer/model"
	"github.com/enkaigaku/dvd-rental/internal/customer/repository"
	"github.com/enkaigaku/dvd-rental/pkg/money"
)

// CustomerService contains business logic for customer operations.
type CustomerService struct {
	customerRepo repository.CustomerRepository
	addressRepo  repository.AddressRepository
	cityRepo     repository.CityRepository
	countryRepo  repository.CountryRepository
}

// NewCustomerService creates a new CustomerService.
func NewCustomerService(
	customerRepo repository.CustomerRepository,
	addressRepo repository.AddressRepository,
	cityRepo repository.CityRepository,
	countryRepo repository.CountryRepository,
) *CustomerService {
	return &CustomerService{
		customerRepo: customerRepo,
		addressRepo:  addressRepo,
		cityRepo:     cityRepo,
		countryRepo:  countryRepo,
	}
}

// GetCustomerByEmail retrieves a customer by email (includes password_hash for auth).
func (s *CustomerService) GetCustomerByEmail(ctx context.Context, email string) (model.Customer, error) {
	if email == "" {
		return model.Customer{}, fmt.Errorf("email must not be empty: %w", ErrInvalidArgument)
	}

	cust, err := s.customerRepo.GetCustomerByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Customer{}, fmt.Errorf("customer with email %q: %w", email, ErrNotFound)
		}
		return model.Customer{}, err
	}

	return cust, nil
}

// GetCustomer returns a customer with enriched address details.
func (s *CustomerService) GetCustomer(ctx context.Context, customerID int32) (model.CustomerDetail, error) {
	if customerID <= 0 {
		return model.CustomerDetail{}, fmt.Errorf("customer_id must be positive: %w", ErrInvalidArgument)
	}

	cust, err := s.customerRepo.GetCustomer(ctx, customerID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.CustomerDetail{}, fmt.Errorf("customer %d: %w", customerID, ErrNotFound)
		}
		return model.CustomerDetail{}, err
	}

	detail := model.CustomerDetail{Customer: cust}

	// Aggregate address → city → country.
	addr, err := s.addressRepo.GetAddress(ctx, cust.AddressID)
	if err == nil {
		detail.Address = addr.Address
		detail.Address2 = addr.Address2
		detail.District = addr.District
		detail.PostalCode = addr.PostalCode
		detail.Phone = addr.Phone

		city, err := s.cityRepo.GetCity(ctx, addr.CityID)
		if err == nil {
			detail.CityName = city.City

			country, err := s.countryRepo.GetCountry(ctx, city.CountryID)
			if err == nil {
				detail.CountryName = country.Country
			}
		}
	}

	return detail, nil
}

// ListCustomers returns a paginated list of customers.
func (s *CustomerService) ListCustomers(ctx context.Context, pageSize, page int32) ([]model.Customer, int64, error) {
	pageSize, page = clampPagination(pageSize, page)
	offset := (page - 1) * pageSize

	customers, err := s.customerRepo.ListCustomers(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.customerRepo.CountCustomers(ctx)
	if err != nil {
		return nil, 0, err
	}

	return customers, total, nil
}

// ListCustomersByStore returns customers belonging to a given store.
func (s *CustomerService) ListCustomersByStore(ctx context.Context, storeID, pageSize, page int32) ([]model.Customer, int64, error) {
	if storeID <= 0 {
		return nil, 0, fmt.Errorf("store_id must be positive: %w", ErrInvalidArgument)
	}

	pageSize, page = clampPagination(pageSize, page)
	offset := (page - 1) * pageSize

	customers, err := s.customerRepo.ListCustomersByStore(ctx, storeID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.customerRepo.CountCustomersByStore(ctx, storeID)
	if err != nil {
		return nil, 0, err
	}

	return customers, total, nil
}

// CreateCustomer creates a new customer after validation.
func (s *CustomerService) CreateCustomer(ctx context.Context, params repository.CreateCustomerParams) (model.Customer, error) {
	if err := s.validateCustomerParams(ctx, params.FirstName, params.LastName, params.StoreID, params.AddressID); err != nil {
		return model.Customer{}, err
	}

	cust, err := s.customerRepo.CreateCustomer(ctx, params)
	if err != nil {
		if isForeignKeyViolation(err) {
			return model.Customer{}, fmt.Errorf("invalid store_id or address_id: %w", ErrInvalidArgument)
		}
		if isUniqueViolation(err) {
			return model.Customer{}, fmt.Errorf("email already in use: %w", ErrAlreadyExists)
		}
		return model.Customer{}, err
	}
	return cust, nil
}

// UpdateCustomer updates an existing customer after validation.
func (s *CustomerService) UpdateCustomer(ctx context.Context, params repository.UpdateCustomerParams) (model.Customer, error) {
	if params.CustomerID <= 0 {
		return model.Customer{}, fmt.Errorf("customer_id must be positive: %w", ErrInvalidArgument)
	}

	if err := s.validateCustomerParams(ctx, params.FirstName, params.LastName, params.StoreID, params.AddressID); err != nil {
		return model.Customer{}, err
	}

	cust, err := s.customerRepo.UpdateCustomer(ctx, params)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Customer{}, fmt.Errorf("customer %d: %w", params.CustomerID, ErrNotFound)
		}
		if isForeignKeyViolation(err) {
			return model.Customer{}, fmt.Errorf("invalid store_id or address_id: %w", ErrInvalidArgument)
		}
		if isUniqueViolation(err) {
			return model.Customer{}, fmt.Errorf("email already in use: %w", ErrAlreadyExists)
		}
		return model.Customer{}, err
	}
	return cust, nil
}

// DeleteCustomer deletes a customer. Fails if rental or payment records reference it.
func (s *CustomerService) DeleteCustomer(ctx context.Context, customerID int32) error {
	if customerID <= 0 {
		return fmt.Errorf("customer_id must be positive: %w", ErrInvalidArgument)
	}

	// Verify the customer exists.
	_, err := s.customerRepo.GetCustomer(ctx, customerID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("customer %d: %w", customerID, ErrNotFound)
		}
		return err
	}

	if err := s.customerRepo.DeleteCustomer(ctx, customerID); err != nil {
		if isForeignKeyViolation(err) {
			return fmt.Errorf("customer %d is referenced by rental or payment records: %w", customerID, ErrForeignKey)
		}
		return err
	}
	return nil
}

func (s *CustomerService) validateCustomerParams(ctx context.Context, firstName, lastName string, storeID, addressID int32) error {
	if firstName == "" {
		return fmt.Errorf("first_name must not be empty: %w", ErrInvalidArgument)
	}
	if lastName == "" {
		return fmt.Errorf("last_name must not be empty: %w", ErrInvalidArgument)
	}
	if storeID <= 0 {
		return fmt.Errorf("store_id must be positive: %w", ErrInvalidArgument)
	}
	if addressID <= 0 {
		return fmt.Errorf("address_id must be positive: %w", ErrInvalidArgument)
	}

	// Verify address exists.
	if _, err := s.addressRepo.GetAddress(ctx, addressID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("address %d not found: %w", addressID, ErrInvalidArgument)
		}
		return err
	}

	return nil
}

// GetCustomerStanding returns account standing info: whether overdue, active rentals, balance.
func (s *CustomerService) GetCustomerStanding(ctx context.Context, customerID int32) (model.CustomerStanding, error) {
	if customerID <= 0 {
		return model.CustomerStanding{}, fmt.Errorf("customer_id must be positive: %w", ErrInvalidArgument)
	}
	customer, err := s.customerRepo.GetCustomer(ctx, customerID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.CustomerStanding{}, fmt.Errorf("customer %d: %w", customerID, ErrNotFound)
		}
		return model.CustomerStanding{}, err
	}

	overdue, err := s.customerRepo.CountCustomerOverdueRentals(ctx, customerID)
	if err != nil {
		return model.CustomerStanding{}, err
	}
	active, err := s.customerRepo.CountCustomerActiveRentals(ctx, customerID)
	if err != nil {
		return model.CustomerStanding{}, err
	}
	balance, err := s.customerRepo.GetCustomerBalance(ctx, customerID)
	if err != nil {
		return model.CustomerStanding{}, err
	}

	reasons := make([]string, 0)
	inGoodStanding := true
	if !customer.Active {
		inGoodStanding = false
		reasons = append(reasons, "account is inactive")
	}
	amount, err := money.Parse(balance)
	if err != nil {
		return model.CustomerStanding{}, err
	}
	balance = amount.FloatString(2)
	if amount.Sign() > 0 {
		inGoodStanding = false
		reasons = append(reasons, "outstanding balance: "+balance)
	}
	if overdue > 0 {
		inGoodStanding = false
		reasons = append(reasons, fmt.Sprintf("%d overdue rental(s)", overdue))
	}
	if active >= 5 {
		inGoodStanding = false
		reasons = append(reasons, "rental limit reached (5 active rentals)")
	}

	return model.CustomerStanding{
		CustomerID:         customerID,
		InGoodStanding:     inGoodStanding,
		Reasons:            reasons,
		ActiveRentals:      int32(active),
		OverdueRentals:     int32(overdue),
		OutstandingBalance: balance,
	}, nil
}

// GetCustomerSummary returns aggregate rental statistics for a customer.
func (s *CustomerService) GetCustomerSummary(ctx context.Context, customerID int32) (model.CustomerSummary, error) {
	if customerID <= 0 {
		return model.CustomerSummary{}, fmt.Errorf("customer_id must be positive: %w", ErrInvalidArgument)
	}
	_, err := s.customerRepo.GetCustomer(ctx, customerID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.CustomerSummary{}, fmt.Errorf("customer %d: %w", customerID, ErrNotFound)
		}
		return model.CustomerSummary{}, err
	}

	totalRentals, err := s.customerRepo.CountCustomerTotalRentals(ctx, customerID)
	if err != nil {
		return model.CustomerSummary{}, err
	}
	active, err := s.customerRepo.CountCustomerActiveRentals(ctx, customerID)
	if err != nil {
		return model.CustomerSummary{}, err
	}
	totalSpent, err := s.customerRepo.GetCustomerTotalSpent(ctx, customerID)
	if err != nil {
		return model.CustomerSummary{}, err
	}
	favoriteCategory, err := s.customerRepo.GetCustomerFavoriteCategory(ctx, customerID)
	if err != nil {
		return model.CustomerSummary{}, err
	}
	balance, err := s.customerRepo.GetCustomerBalance(ctx, customerID)
	if err != nil {
		return model.CustomerSummary{}, err
	}

	return model.CustomerSummary{
		CustomerID:         customerID,
		TotalRentals:       int32(totalRentals),
		ActiveRentals:      int32(active),
		TotalSpent:         totalSpent,
		FavoriteCategory:   favoriteCategory,
		OutstandingBalance: balance,
	}, nil
}
