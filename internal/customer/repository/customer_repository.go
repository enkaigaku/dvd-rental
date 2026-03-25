package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/enkaigaku/dvd-rental/internal/customer/model"
	"github.com/enkaigaku/dvd-rental/gen/sqlc/customer"
)

// ErrNotFound is returned when a queried entity does not exist.
var ErrNotFound = errors.New("not found")

// CreateCustomerParams holds parameters for creating a customer.
type CreateCustomerParams struct {
	StoreID   int32
	FirstName string
	LastName  string
	Email     string
	AddressID int32
	Active    bool
}

// UpdateCustomerParams holds parameters for updating a customer.
type UpdateCustomerParams struct {
	CustomerID int32
	StoreID    int32
	FirstName  string
	LastName   string
	Email      string
	AddressID  int32
	Active     bool
}

// CustomerRepository defines data-access operations for customers.
type CustomerRepository interface {
	GetCustomer(ctx context.Context, customerID int32) (model.Customer, error)
	GetCustomerByEmail(ctx context.Context, email string) (model.Customer, error)
	ListCustomers(ctx context.Context, limit, offset int32) ([]model.Customer, error)
	CountCustomers(ctx context.Context) (int64, error)
	ListCustomersByStore(ctx context.Context, storeID, limit, offset int32) ([]model.Customer, error)
	CountCustomersByStore(ctx context.Context, storeID int32) (int64, error)
	CreateCustomer(ctx context.Context, params CreateCustomerParams) (model.Customer, error)
	UpdateCustomer(ctx context.Context, params UpdateCustomerParams) (model.Customer, error)
	DeleteCustomer(ctx context.Context, customerID int32) error
	CountCustomerOverdueRentals(ctx context.Context, customerID int32) (int64, error)
	CountCustomerActiveRentals(ctx context.Context, customerID int32) (int64, error)
	GetCustomerBalance(ctx context.Context, customerID int32) (string, error)
	CountCustomerTotalRentals(ctx context.Context, customerID int32) (int64, error)
	GetCustomerTotalSpent(ctx context.Context, customerID int32) (string, error)
	GetCustomerFavoriteCategory(ctx context.Context, customerID int32) (string, error)
}

type customerRepository struct {
	q *customersqlc.Queries
}

// NewCustomerRepository creates a new CustomerRepository.
func NewCustomerRepository(pool *pgxpool.Pool) CustomerRepository {
	return &customerRepository{q: customersqlc.New(pool)}
}

func (r *customerRepository) GetCustomer(ctx context.Context, customerID int32) (model.Customer, error) {
	row, err := r.q.GetCustomer(ctx, customerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Customer{}, ErrNotFound
		}
		return model.Customer{}, fmt.Errorf("get customer: %w", err)
	}
	return toCustomerModel(row.CustomerID, row.StoreID, row.FirstName, row.LastName,
		row.Email, row.AddressID, row.Activebool, row.CreateDate, row.LastUpdate), nil
}

func (r *customerRepository) GetCustomerByEmail(ctx context.Context, email string) (model.Customer, error) {
	row, err := r.q.GetCustomerByEmail(ctx, stringToText(email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Customer{}, ErrNotFound
		}
		return model.Customer{}, fmt.Errorf("get customer by email: %w", err)
	}
	c := toCustomerModel(row.CustomerID, row.StoreID, row.FirstName, row.LastName,
		row.Email, row.AddressID, row.Activebool, row.CreateDate, row.LastUpdate)
	c.PasswordHash = textToString(row.PasswordHash)
	return c, nil
}

func (r *customerRepository) ListCustomers(ctx context.Context, limit, offset int32) ([]model.Customer, error) {
	rows, err := r.q.ListCustomers(ctx, customersqlc.ListCustomersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list customers: %w", err)
	}
	customers := make([]model.Customer, len(rows))
	for i, row := range rows {
		customers[i] = toCustomerModel(row.CustomerID, row.StoreID, row.FirstName, row.LastName,
			row.Email, row.AddressID, row.Activebool, row.CreateDate, row.LastUpdate)
	}
	return customers, nil
}

func (r *customerRepository) CountCustomers(ctx context.Context) (int64, error) {
	count, err := r.q.CountCustomers(ctx)
	if err != nil {
		return 0, fmt.Errorf("count customers: %w", err)
	}
	return count, nil
}

func (r *customerRepository) ListCustomersByStore(ctx context.Context, storeID, limit, offset int32) ([]model.Customer, error) {
	rows, err := r.q.ListCustomersByStore(ctx, customersqlc.ListCustomersByStoreParams{
		StoreID: storeID,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list customers by store: %w", err)
	}
	customers := make([]model.Customer, len(rows))
	for i, row := range rows {
		customers[i] = toCustomerModel(row.CustomerID, row.StoreID, row.FirstName, row.LastName,
			row.Email, row.AddressID, row.Activebool, row.CreateDate, row.LastUpdate)
	}
	return customers, nil
}

func (r *customerRepository) CountCustomersByStore(ctx context.Context, storeID int32) (int64, error) {
	count, err := r.q.CountCustomersByStore(ctx, storeID)
	if err != nil {
		return 0, fmt.Errorf("count customers by store: %w", err)
	}
	return count, nil
}

func (r *customerRepository) CreateCustomer(ctx context.Context, params CreateCustomerParams) (model.Customer, error) {
	row, err := r.q.CreateCustomer(ctx, customersqlc.CreateCustomerParams{
		StoreID:    params.StoreID,
		FirstName:  params.FirstName,
		LastName:   params.LastName,
		Email:      stringToText(params.Email),
		AddressID:  params.AddressID,
		Activebool: params.Active,
		Active:     boolToActive(params.Active),
	})
	if err != nil {
		return model.Customer{}, fmt.Errorf("create customer: %w", err)
	}
	return toCustomerModel(row.CustomerID, row.StoreID, row.FirstName, row.LastName,
		row.Email, row.AddressID, row.Activebool, row.CreateDate, row.LastUpdate), nil
}

func (r *customerRepository) UpdateCustomer(ctx context.Context, params UpdateCustomerParams) (model.Customer, error) {
	row, err := r.q.UpdateCustomer(ctx, customersqlc.UpdateCustomerParams{
		CustomerID: params.CustomerID,
		StoreID:    params.StoreID,
		FirstName:  params.FirstName,
		LastName:   params.LastName,
		Email:      stringToText(params.Email),
		AddressID:  params.AddressID,
		Activebool: params.Active,
		Active:     boolToActive(params.Active),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Customer{}, ErrNotFound
		}
		return model.Customer{}, fmt.Errorf("update customer: %w", err)
	}
	return toCustomerModel(row.CustomerID, row.StoreID, row.FirstName, row.LastName,
		row.Email, row.AddressID, row.Activebool, row.CreateDate, row.LastUpdate), nil
}

func (r *customerRepository) DeleteCustomer(ctx context.Context, customerID int32) error {
	if err := r.q.DeleteCustomer(ctx, customerID); err != nil {
		return fmt.Errorf("delete customer: %w", err)
	}
	return nil
}

func (r *customerRepository) CountCustomerOverdueRentals(ctx context.Context, customerID int32) (int64, error) {
	count, err := r.q.CountCustomerOverdueRentals(ctx, customerID)
	if err != nil {
		return 0, fmt.Errorf("count overdue rentals: %w", err)
	}
	return count, nil
}

func (r *customerRepository) CountCustomerActiveRentals(ctx context.Context, customerID int32) (int64, error) {
	count, err := r.q.CountCustomerActiveRentals(ctx, customerID)
	if err != nil {
		return 0, fmt.Errorf("count active rentals: %w", err)
	}
	return count, nil
}

func (r *customerRepository) GetCustomerBalance(ctx context.Context, customerID int32) (string, error) {
	val, err := r.q.GetCustomerBalance(ctx, customerID)
	if err != nil {
		return "0.00", fmt.Errorf("get customer balance: %w", err)
	}
	if s, ok := val.(string); ok {
		return s, nil
	}
	// numeric result may come back as pgtype.Numeric; format it
	if f, ok := numericToFloat(val); ok {
		return fmt.Sprintf("%.2f", f), nil
	}
	return "0.00", nil
}

func (r *customerRepository) CountCustomerTotalRentals(ctx context.Context, customerID int32) (int64, error) {
	count, err := r.q.CountCustomerTotalRentals(ctx, customerID)
	if err != nil {
		return 0, fmt.Errorf("count total rentals: %w", err)
	}
	return count, nil
}

func (r *customerRepository) GetCustomerTotalSpent(ctx context.Context, customerID int32) (string, error) {
	val, err := r.q.GetCustomerTotalSpent(ctx, customerID)
	if err != nil {
		return "0.00", fmt.Errorf("get customer total spent: %w", err)
	}
	if s, ok := val.(string); ok {
		return s, nil
	}
	if f, ok := numericToFloat(val); ok {
		return fmt.Sprintf("%.2f", f), nil
	}
	return "0.00", nil
}

func (r *customerRepository) GetCustomerFavoriteCategory(ctx context.Context, customerID int32) (string, error) {
	name, err := r.q.GetCustomerFavoriteCategory(ctx, customerID)
	if err != nil {
		return "", nil // no favorite category is not an error
	}
	return name, nil
}

func toCustomerModel(
	customerID, storeID int32,
	firstName, lastName string,
	email pgtype.Text,
	addressID int32,
	activebool bool,
	createDate pgtype.Date,
	lastUpdate pgtype.Timestamptz,
) model.Customer {
	return model.Customer{
		CustomerID: customerID,
		StoreID:    storeID,
		FirstName:  firstName,
		LastName:   lastName,
		Email:      textToString(email),
		AddressID:  addressID,
		Active:     activebool,
		CreateDate: dateToTime(createDate),
		LastUpdate: timestamptzToTime(lastUpdate),
	}
}
