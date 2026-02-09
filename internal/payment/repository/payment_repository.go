package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/enkaigaku/dvd-rental/internal/payment/model"
	"github.com/enkaigaku/dvd-rental/internal/payment/repository/sqlcgen"
)

// ErrNotFound is returned when a queried entity does not exist.
var ErrNotFound = errors.New("not found")

// CreatePaymentParams holds parameters for creating a payment.
type CreatePaymentParams struct {
	CustomerID int32
	StaffID    int32
	RentalID   int32
	Amount     string
}

// PaymentRepository defines data-access operations for payments.
type PaymentRepository interface {
	GetPayment(ctx context.Context, paymentID int32) (model.Payment, error)
	ListPayments(ctx context.Context, limit, offset int32) ([]model.Payment, error)
	CountPayments(ctx context.Context) (int64, error)
	ListPaymentsByCustomer(ctx context.Context, customerID, limit, offset int32) ([]model.Payment, error)
	CountPaymentsByCustomer(ctx context.Context, customerID int32) (int64, error)
	ListPaymentsByStaff(ctx context.Context, staffID, limit, offset int32) ([]model.Payment, error)
	CountPaymentsByStaff(ctx context.Context, staffID int32) (int64, error)
	ListPaymentsByRental(ctx context.Context, rentalID, limit, offset int32) ([]model.Payment, error)
	CountPaymentsByRental(ctx context.Context, rentalID int32) (int64, error)
	ListPaymentsByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int32) ([]model.Payment, error)
	CountPaymentsByDateRange(ctx context.Context, startDate, endDate time.Time) (int64, error)
	CreatePayment(ctx context.Context, params CreatePaymentParams) (model.Payment, error)
	DeletePayment(ctx context.Context, paymentID int32) error
	GetCustomerName(ctx context.Context, customerID int32) (string, error)
	GetStaffName(ctx context.Context, staffID int32) (string, error)
	GetRentalDate(ctx context.Context, rentalID int32) (time.Time, error)
}

type paymentRepository struct {
	q *sqlcgen.Queries
}

// NewPaymentRepository creates a new PaymentRepository.
func NewPaymentRepository(pool *pgxpool.Pool) PaymentRepository {
	return &paymentRepository{q: sqlcgen.New(pool)}
}

func (r *paymentRepository) GetPayment(ctx context.Context, paymentID int32) (model.Payment, error) {
	row, err := r.q.GetPayment(ctx, paymentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Payment{}, ErrNotFound
		}
		return model.Payment{}, fmt.Errorf("get payment: %w", err)
	}
	return toPaymentModel(row), nil
}

func (r *paymentRepository) ListPayments(ctx context.Context, limit, offset int32) ([]model.Payment, error) {
	rows, err := r.q.ListPayments(ctx, sqlcgen.ListPaymentsParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("list payments: %w", err)
	}
	return toPaymentModels(rows), nil
}

func (r *paymentRepository) CountPayments(ctx context.Context) (int64, error) {
	count, err := r.q.CountPayments(ctx)
	if err != nil {
		return 0, fmt.Errorf("count payments: %w", err)
	}
	return count, nil
}

func (r *paymentRepository) ListPaymentsByCustomer(ctx context.Context, customerID, limit, offset int32) ([]model.Payment, error) {
	rows, err := r.q.ListPaymentsByCustomer(ctx, sqlcgen.ListPaymentsByCustomerParams{
		CustomerID: customerID, Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list payments by customer: %w", err)
	}
	return toPaymentModels(rows), nil
}

func (r *paymentRepository) CountPaymentsByCustomer(ctx context.Context, customerID int32) (int64, error) {
	count, err := r.q.CountPaymentsByCustomer(ctx, customerID)
	if err != nil {
		return 0, fmt.Errorf("count payments by customer: %w", err)
	}
	return count, nil
}

func (r *paymentRepository) ListPaymentsByStaff(ctx context.Context, staffID, limit, offset int32) ([]model.Payment, error) {
	rows, err := r.q.ListPaymentsByStaff(ctx, sqlcgen.ListPaymentsByStaffParams{
		StaffID: staffID, Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list payments by staff: %w", err)
	}
	return toPaymentModels(rows), nil
}

func (r *paymentRepository) CountPaymentsByStaff(ctx context.Context, staffID int32) (int64, error) {
	count, err := r.q.CountPaymentsByStaff(ctx, staffID)
	if err != nil {
		return 0, fmt.Errorf("count payments by staff: %w", err)
	}
	return count, nil
}

func (r *paymentRepository) ListPaymentsByRental(ctx context.Context, rentalID, limit, offset int32) ([]model.Payment, error) {
	rows, err := r.q.ListPaymentsByRental(ctx, sqlcgen.ListPaymentsByRentalParams{
		RentalID: rentalID, Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list payments by rental: %w", err)
	}
	return toPaymentModels(rows), nil
}

func (r *paymentRepository) CountPaymentsByRental(ctx context.Context, rentalID int32) (int64, error) {
	count, err := r.q.CountPaymentsByRental(ctx, rentalID)
	if err != nil {
		return 0, fmt.Errorf("count payments by rental: %w", err)
	}
	return count, nil
}

func (r *paymentRepository) ListPaymentsByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int32) ([]model.Payment, error) {
	rows, err := r.q.ListPaymentsByDateRange(ctx, sqlcgen.ListPaymentsByDateRangeParams{
		PaymentDate:   timeToTimestamptz(startDate),
		PaymentDate_2: timeToTimestamptz(endDate),
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list payments by date range: %w", err)
	}
	return toPaymentModels(rows), nil
}

func (r *paymentRepository) CountPaymentsByDateRange(ctx context.Context, startDate, endDate time.Time) (int64, error) {
	count, err := r.q.CountPaymentsByDateRange(ctx, sqlcgen.CountPaymentsByDateRangeParams{
		PaymentDate:   timeToTimestamptz(startDate),
		PaymentDate_2: timeToTimestamptz(endDate),
	})
	if err != nil {
		return 0, fmt.Errorf("count payments by date range: %w", err)
	}
	return count, nil
}

func (r *paymentRepository) CreatePayment(ctx context.Context, params CreatePaymentParams) (model.Payment, error) {
	row, err := r.q.CreatePayment(ctx, sqlcgen.CreatePaymentParams{
		CustomerID: params.CustomerID,
		StaffID:    params.StaffID,
		RentalID:   params.RentalID,
		Amount:     stringToNumeric(params.Amount),
	})
	if err != nil {
		return model.Payment{}, fmt.Errorf("create payment: %w", err)
	}
	return toPaymentModel(row), nil
}

func (r *paymentRepository) DeletePayment(ctx context.Context, paymentID int32) error {
	if err := r.q.DeletePayment(ctx, paymentID); err != nil {
		return fmt.Errorf("delete payment: %w", err)
	}
	return nil
}

func (r *paymentRepository) GetCustomerName(ctx context.Context, customerID int32) (string, error) {
	val, err := r.q.GetCustomerName(ctx, customerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("get customer name: %w", err)
	}
	if s, ok := val.(string); ok {
		return s, nil
	}
	return "", nil
}

func (r *paymentRepository) GetStaffName(ctx context.Context, staffID int32) (string, error) {
	val, err := r.q.GetStaffName(ctx, staffID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("get staff name: %w", err)
	}
	if s, ok := val.(string); ok {
		return s, nil
	}
	return "", nil
}

func (r *paymentRepository) GetRentalDate(ctx context.Context, rentalID int32) (time.Time, error) {
	ts, err := r.q.GetRentalDate(ctx, rentalID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, ErrNotFound
		}
		return time.Time{}, fmt.Errorf("get rental date: %w", err)
	}
	return timestamptzToTime(ts), nil
}

func toPaymentModel(p sqlcgen.Payment) model.Payment {
	return model.Payment{
		PaymentID:   p.PaymentID,
		CustomerID:  p.CustomerID,
		StaffID:     p.StaffID,
		RentalID:    p.RentalID,
		Amount:      numericToString(p.Amount),
		PaymentDate: timestamptzToTime(p.PaymentDate),
	}
}

func toPaymentModels(rows []sqlcgen.Payment) []model.Payment {
	payments := make([]model.Payment, len(rows))
	for i, row := range rows {
		payments[i] = toPaymentModel(row)
	}
	return payments
}
