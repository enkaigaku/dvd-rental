package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/enkaigaku/dvd-rental/internal/rental/model"
	"github.com/enkaigaku/dvd-rental/internal/rental/repository/sqlcgen"
)

// ErrNotFound is returned when a queried entity does not exist.
var ErrNotFound = errors.New("not found")

// CreateRentalParams holds parameters for creating a rental.
type CreateRentalParams struct {
	InventoryID int32
	CustomerID  int32
	StaffID     int32
}

// RentalRepository defines data-access operations for rentals.
type RentalRepository interface {
	GetRental(ctx context.Context, rentalID int32) (model.Rental, error)
	ListRentals(ctx context.Context, limit, offset int32) ([]model.Rental, error)
	CountRentals(ctx context.Context) (int64, error)
	ListRentalsByCustomer(ctx context.Context, customerID, limit, offset int32) ([]model.Rental, error)
	CountRentalsByCustomer(ctx context.Context, customerID int32) (int64, error)
	ListRentalsByInventory(ctx context.Context, inventoryID, limit, offset int32) ([]model.Rental, error)
	CountRentalsByInventory(ctx context.Context, inventoryID int32) (int64, error)
	ListOverdueRentals(ctx context.Context, limit, offset int32) ([]model.Rental, error)
	CountOverdueRentals(ctx context.Context) (int64, error)
	CreateRental(ctx context.Context, params CreateRentalParams) (model.Rental, error)
	ReturnRental(ctx context.Context, rentalID int32) (model.Rental, error)
	DeleteRental(ctx context.Context, rentalID int32) error
	GetCustomerName(ctx context.Context, customerID int32) (string, error)
	GetFilmTitleByInventory(ctx context.Context, inventoryID int32) (string, int32, error)
	IsInventoryAvailable(ctx context.Context, inventoryID int32) (bool, error)
}

type rentalRepository struct {
	q *sqlcgen.Queries
}

// NewRentalRepository creates a new RentalRepository.
func NewRentalRepository(pool *pgxpool.Pool) RentalRepository {
	return &rentalRepository{q: sqlcgen.New(pool)}
}

func (r *rentalRepository) GetRental(ctx context.Context, rentalID int32) (model.Rental, error) {
	row, err := r.q.GetRental(ctx, rentalID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Rental{}, ErrNotFound
		}
		return model.Rental{}, fmt.Errorf("get rental: %w", err)
	}
	return toRentalModel(row), nil
}

func (r *rentalRepository) ListRentals(ctx context.Context, limit, offset int32) ([]model.Rental, error) {
	rows, err := r.q.ListRentals(ctx, sqlcgen.ListRentalsParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("list rentals: %w", err)
	}
	return toRentalModels(rows), nil
}

func (r *rentalRepository) CountRentals(ctx context.Context) (int64, error) {
	count, err := r.q.CountRentals(ctx)
	if err != nil {
		return 0, fmt.Errorf("count rentals: %w", err)
	}
	return count, nil
}

func (r *rentalRepository) ListRentalsByCustomer(ctx context.Context, customerID, limit, offset int32) ([]model.Rental, error) {
	rows, err := r.q.ListRentalsByCustomer(ctx, sqlcgen.ListRentalsByCustomerParams{
		CustomerID: customerID, Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list rentals by customer: %w", err)
	}
	return toRentalModels(rows), nil
}

func (r *rentalRepository) CountRentalsByCustomer(ctx context.Context, customerID int32) (int64, error) {
	count, err := r.q.CountRentalsByCustomer(ctx, customerID)
	if err != nil {
		return 0, fmt.Errorf("count rentals by customer: %w", err)
	}
	return count, nil
}

func (r *rentalRepository) ListRentalsByInventory(ctx context.Context, inventoryID, limit, offset int32) ([]model.Rental, error) {
	rows, err := r.q.ListRentalsByInventory(ctx, sqlcgen.ListRentalsByInventoryParams{
		InventoryID: inventoryID, Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list rentals by inventory: %w", err)
	}
	return toRentalModels(rows), nil
}

func (r *rentalRepository) CountRentalsByInventory(ctx context.Context, inventoryID int32) (int64, error) {
	count, err := r.q.CountRentalsByInventory(ctx, inventoryID)
	if err != nil {
		return 0, fmt.Errorf("count rentals by inventory: %w", err)
	}
	return count, nil
}

func (r *rentalRepository) ListOverdueRentals(ctx context.Context, limit, offset int32) ([]model.Rental, error) {
	rows, err := r.q.ListOverdueRentals(ctx, sqlcgen.ListOverdueRentalsParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("list overdue rentals: %w", err)
	}
	return toRentalModels(rows), nil
}

func (r *rentalRepository) CountOverdueRentals(ctx context.Context) (int64, error) {
	count, err := r.q.CountOverdueRentals(ctx)
	if err != nil {
		return 0, fmt.Errorf("count overdue rentals: %w", err)
	}
	return count, nil
}

func (r *rentalRepository) CreateRental(ctx context.Context, params CreateRentalParams) (model.Rental, error) {
	row, err := r.q.CreateRental(ctx, sqlcgen.CreateRentalParams{
		InventoryID: params.InventoryID,
		CustomerID:  params.CustomerID,
		StaffID:     params.StaffID,
	})
	if err != nil {
		return model.Rental{}, fmt.Errorf("create rental: %w", err)
	}
	return toRentalModel(row), nil
}

func (r *rentalRepository) ReturnRental(ctx context.Context, rentalID int32) (model.Rental, error) {
	row, err := r.q.ReturnRental(ctx, rentalID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Rental{}, ErrNotFound
		}
		return model.Rental{}, fmt.Errorf("return rental: %w", err)
	}
	return toRentalModel(row), nil
}

func (r *rentalRepository) DeleteRental(ctx context.Context, rentalID int32) error {
	if err := r.q.DeleteRental(ctx, rentalID); err != nil {
		return fmt.Errorf("delete rental: %w", err)
	}
	return nil
}

func (r *rentalRepository) GetCustomerName(ctx context.Context, customerID int32) (string, error) {
	val, err := r.q.GetCustomerName(ctx, customerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("get customer name: %w", err)
	}
	// sqlc maps the concatenation expression to interface{}.
	if s, ok := val.(string); ok {
		return s, nil
	}
	return "", nil
}

func (r *rentalRepository) GetFilmTitleByInventory(ctx context.Context, inventoryID int32) (string, int32, error) {
	row, err := r.q.GetFilmTitleByInventory(ctx, inventoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", 0, ErrNotFound
		}
		return "", 0, fmt.Errorf("get film title by inventory: %w", err)
	}
	return row.Title, row.StoreID, nil
}

func (r *rentalRepository) IsInventoryAvailable(ctx context.Context, inventoryID int32) (bool, error) {
	available, err := r.q.IsInventoryAvailable(ctx, inventoryID)
	if err != nil {
		return false, fmt.Errorf("is inventory available: %w", err)
	}
	return available, nil
}

func toRentalModel(r sqlcgen.Rental) model.Rental {
	return model.Rental{
		RentalID:    r.RentalID,
		RentalDate:  timestamptzToTime(r.RentalDate),
		InventoryID: r.InventoryID,
		CustomerID:  r.CustomerID,
		ReturnDate:  timestamptzToTime(r.ReturnDate),
		StaffID:     r.StaffID,
		LastUpdate:  timestamptzToTime(r.LastUpdate),
	}
}

func toRentalModels(rows []sqlcgen.Rental) []model.Rental {
	rentals := make([]model.Rental, len(rows))
	for i, row := range rows {
		rentals[i] = toRentalModel(row)
	}
	return rentals
}
