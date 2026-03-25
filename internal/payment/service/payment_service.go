package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/enkaigaku/dvd-rental/internal/payment/model"
	"github.com/enkaigaku/dvd-rental/internal/payment/repository"
)

// PaymentService contains business logic for payment operations.
type PaymentService struct {
	repo repository.PaymentRepository
}

// NewPaymentService creates a new PaymentService.
func NewPaymentService(repo repository.PaymentRepository) *PaymentService {
	return &PaymentService{repo: repo}
}

// GetPayment returns a payment with enriched details (customer name, staff name, rental date).
func (s *PaymentService) GetPayment(ctx context.Context, paymentID int32) (model.PaymentDetail, error) {
	if paymentID <= 0 {
		return model.PaymentDetail{}, fmt.Errorf("payment_id must be positive: %w", ErrInvalidArgument)
	}

	payment, err := s.repo.GetPayment(ctx, paymentID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.PaymentDetail{}, fmt.Errorf("payment %d: %w", paymentID, ErrNotFound)
		}
		return model.PaymentDetail{}, err
	}

	detail := model.PaymentDetail{Payment: payment}

	// Aggregate customer name.
	name, err := s.repo.GetCustomerName(ctx, payment.CustomerID)
	if err == nil {
		detail.CustomerName = name
	}

	// Aggregate staff name.
	staffName, err := s.repo.GetStaffName(ctx, payment.StaffID)
	if err == nil {
		detail.StaffName = staffName
	}

	// Aggregate rental date.
	rentalDate, err := s.repo.GetRentalDate(ctx, payment.RentalID)
	if err == nil {
		detail.RentalDate = rentalDate
	}

	return detail, nil
}

// ListPayments returns a paginated list of payments.
func (s *PaymentService) ListPayments(ctx context.Context, pageSize, page int32) ([]model.Payment, int64, error) {
	pageSize, page = clampPagination(pageSize, page)
	offset := (page - 1) * pageSize

	payments, err := s.repo.ListPayments(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.CountPayments(ctx)
	if err != nil {
		return nil, 0, err
	}

	return payments, total, nil
}

// ListPaymentsByCustomer returns payments for a given customer.
func (s *PaymentService) ListPaymentsByCustomer(ctx context.Context, customerID, pageSize, page int32) ([]model.Payment, int64, error) {
	if customerID <= 0 {
		return nil, 0, fmt.Errorf("customer_id must be positive: %w", ErrInvalidArgument)
	}

	pageSize, page = clampPagination(pageSize, page)
	offset := (page - 1) * pageSize

	payments, err := s.repo.ListPaymentsByCustomer(ctx, customerID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.CountPaymentsByCustomer(ctx, customerID)
	if err != nil {
		return nil, 0, err
	}

	return payments, total, nil
}

// ListPaymentsByStaff returns payments processed by a given staff member.
func (s *PaymentService) ListPaymentsByStaff(ctx context.Context, staffID, pageSize, page int32) ([]model.Payment, int64, error) {
	if staffID <= 0 {
		return nil, 0, fmt.Errorf("staff_id must be positive: %w", ErrInvalidArgument)
	}

	pageSize, page = clampPagination(pageSize, page)
	offset := (page - 1) * pageSize

	payments, err := s.repo.ListPaymentsByStaff(ctx, staffID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.CountPaymentsByStaff(ctx, staffID)
	if err != nil {
		return nil, 0, err
	}

	return payments, total, nil
}

// ListPaymentsByRental returns payments for a given rental.
func (s *PaymentService) ListPaymentsByRental(ctx context.Context, rentalID, pageSize, page int32) ([]model.Payment, int64, error) {
	if rentalID <= 0 {
		return nil, 0, fmt.Errorf("rental_id must be positive: %w", ErrInvalidArgument)
	}

	pageSize, page = clampPagination(pageSize, page)
	offset := (page - 1) * pageSize

	payments, err := s.repo.ListPaymentsByRental(ctx, rentalID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.CountPaymentsByRental(ctx, rentalID)
	if err != nil {
		return nil, 0, err
	}

	return payments, total, nil
}

// ListPaymentsByDateRange returns payments within a date range [startDate, endDate).
func (s *PaymentService) ListPaymentsByDateRange(ctx context.Context, startDate, endDate time.Time, pageSize, page int32) ([]model.Payment, int64, error) {
	if startDate.IsZero() {
		return nil, 0, fmt.Errorf("start_date must not be empty: %w", ErrInvalidArgument)
	}
	if endDate.IsZero() {
		return nil, 0, fmt.Errorf("end_date must not be empty: %w", ErrInvalidArgument)
	}
	if !startDate.Before(endDate) {
		return nil, 0, fmt.Errorf("start_date must be before end_date: %w", ErrInvalidArgument)
	}

	pageSize, page = clampPagination(pageSize, page)
	offset := (page - 1) * pageSize

	payments, err := s.repo.ListPaymentsByDateRange(ctx, startDate, endDate, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.CountPaymentsByDateRange(ctx, startDate, endDate)
	if err != nil {
		return nil, 0, err
	}

	return payments, total, nil
}

// CreatePayment creates a new payment after validation.
func (s *PaymentService) CreatePayment(ctx context.Context, params repository.CreatePaymentParams) (model.Payment, error) {
	if params.CustomerID <= 0 {
		return model.Payment{}, fmt.Errorf("customer_id must be positive: %w", ErrInvalidArgument)
	}
	if params.StaffID <= 0 {
		return model.Payment{}, fmt.Errorf("staff_id must be positive: %w", ErrInvalidArgument)
	}
	if params.RentalID <= 0 {
		return model.Payment{}, fmt.Errorf("rental_id must be positive: %w", ErrInvalidArgument)
	}
	if params.Amount == "" {
		return model.Payment{}, fmt.Errorf("amount must not be empty: %w", ErrInvalidArgument)
	}

	payment, err := s.repo.CreatePayment(ctx, params)
	if err != nil {
		if isForeignKeyViolation(err) {
			return model.Payment{}, fmt.Errorf("invalid customer_id, staff_id, or rental_id: %w", ErrInvalidArgument)
		}
		return model.Payment{}, err
	}
	return payment, nil
}

// GetCustomerBalance returns the financial summary for a customer.
func (s *PaymentService) GetCustomerBalance(ctx context.Context, customerID int32) (model.CustomerBalance, error) {
	if customerID <= 0 {
		return model.CustomerBalance{}, fmt.Errorf("customer_id must be positive: %w", ErrInvalidArgument)
	}

	totalCharges, err := s.repo.GetCustomerTotalCharges(ctx, customerID)
	if err != nil {
		return model.CustomerBalance{}, err
	}

	totalPayments, err := s.repo.GetCustomerTotalPayments(ctx, customerID)
	if err != nil {
		return model.CustomerBalance{}, err
	}

	// Compute balance = charges - payments using strconv arithmetic.
	charges := parseDecimal(totalCharges)
	payments := parseDecimal(totalPayments)
	balance := charges - payments

	rentalCount, _ := s.repo.CountRentalsByCustomer(ctx, customerID)
	paymentCount, _ := s.repo.CountPaymentsByCustomer(ctx, customerID)

	return model.CustomerBalance{
		CustomerID:    customerID,
		TotalCharges:  totalCharges,
		TotalPayments: totalPayments,
		Balance:       fmt.Sprintf("%.2f", balance),
		RentalCount:   int32(rentalCount),
		PaymentCount:  int32(paymentCount),
	}, nil
}

// GetRevenueByStore returns aggregated revenue by store for a date range.
func (s *PaymentService) GetRevenueByStore(ctx context.Context, startDate, endDate time.Time) ([]model.StoreRevenue, string, error) {
	if startDate.IsZero() {
		return nil, "", fmt.Errorf("start_date must not be empty: %w", ErrInvalidArgument)
	}
	if endDate.IsZero() {
		return nil, "", fmt.Errorf("end_date must not be empty: %w", ErrInvalidArgument)
	}
	if !startDate.Before(endDate) {
		return nil, "", fmt.Errorf("start_date must be before end_date: %w", ErrInvalidArgument)
	}

	stores, err := s.repo.GetStoreRevenue(ctx, startDate, endDate)
	if err != nil {
		return nil, "", err
	}

	var total float64
	for _, s := range stores {
		total += parseDecimal(s.TotalRevenue)
	}

	return stores, fmt.Sprintf("%.2f", total), nil
}

func parseDecimal(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

// DeletePayment deletes a payment record.
func (s *PaymentService) DeletePayment(ctx context.Context, paymentID int32) error {
	if paymentID <= 0 {
		return fmt.Errorf("payment_id must be positive: %w", ErrInvalidArgument)
	}

	// Verify the payment exists.
	_, err := s.repo.GetPayment(ctx, paymentID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("payment %d: %w", paymentID, ErrNotFound)
		}
		return err
	}

	if err := s.repo.DeletePayment(ctx, paymentID); err != nil {
		return err
	}
	return nil
}
