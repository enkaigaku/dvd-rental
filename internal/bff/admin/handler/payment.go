package handler

import (
	"context"
	"net/http"
	"time"

	paymentv1 "github.com/enkaigaku/dvd-rental/gen/proto/payment/v1"
)

// PaymentHandler handles payment management endpoints.
type PaymentHandler struct {
	paymentClient paymentv1.PaymentServiceClient
}

// NewPaymentHandler creates a new PaymentHandler.
func NewPaymentHandler(paymentClient paymentv1.PaymentServiceClient) *PaymentHandler {
	return &PaymentHandler{paymentClient: paymentClient}
}

// --- JSON models ---

type paymentResponse struct {
	PaymentID   int32  `json:"payment_id"`
	CustomerID  int32  `json:"customer_id"`
	StaffID     int32  `json:"staff_id"`
	RentalID    int32  `json:"rental_id"`
	Amount      string `json:"amount"`
	PaymentDate string `json:"payment_date"`
}

type paymentDetailResponse struct {
	paymentResponse
	CustomerName string `json:"customer_name"`
	StaffName    string `json:"staff_name"`
	RentalDate   string `json:"rental_date,omitempty"`
}

type paymentListResponse struct {
	Payments   []paymentResponse `json:"payments"`
	TotalCount int32             `json:"total_count"`
}

type createPaymentRequest struct {
	CustomerID int32  `json:"customer_id"`
	StaffID    int32  `json:"staff_id"`
	RentalID   int32  `json:"rental_id"`
	Amount     string `json:"amount"`
}

func paymentToResponse(p *paymentv1.Payment) paymentResponse {
	return paymentResponse{
		PaymentID:   p.GetPaymentId(),
		CustomerID:  p.GetCustomerId(),
		StaffID:     p.GetStaffId(),
		RentalID:    p.GetRentalId(),
		Amount:      p.GetAmount(),
		PaymentDate: p.GetPaymentDate().AsTime().Format(time.RFC3339),
	}
}

func paymentDetailToResponse(d *paymentv1.PaymentDetail) paymentDetailResponse {
	resp := paymentDetailResponse{
		paymentResponse: paymentToResponse(d.GetPayment()),
		CustomerName:    d.GetCustomerName(),
		StaffName:       d.GetStaffName(),
	}
	if d.GetRentalDate() != nil {
		resp.RentalDate = d.GetRentalDate().AsTime().Format(time.RFC3339)
	}
	return resp
}

// ListPayments returns a paginated list of payments.
func (h *PaymentHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	pageSize, page := parsePagination(r)
	customerID := parseQueryInt32(r, "customer_id")
	staffID := parseQueryInt32(r, "staff_id")
	rentalID := parseQueryInt32(r, "rental_id")

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var resp *paymentv1.ListPaymentsResponse
	var err error

	switch {
	case customerID > 0:
		resp, err = h.paymentClient.ListPaymentsByCustomer(ctx, &paymentv1.ListPaymentsByCustomerRequest{
			CustomerId: customerID,
			PageSize:   pageSize,
			Page:       page,
		})
	case staffID > 0:
		resp, err = h.paymentClient.ListPaymentsByStaff(ctx, &paymentv1.ListPaymentsByStaffRequest{
			StaffId:  staffID,
			PageSize: pageSize,
			Page:     page,
		})
	case rentalID > 0:
		resp, err = h.paymentClient.ListPaymentsByRental(ctx, &paymentv1.ListPaymentsByRentalRequest{
			RentalId: rentalID,
			PageSize: pageSize,
			Page:     page,
		})
	default:
		resp, err = h.paymentClient.ListPayments(ctx, &paymentv1.ListPaymentsRequest{
			PageSize: pageSize,
			Page:     page,
		})
	}
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	payments := make([]paymentResponse, len(resp.GetPayments()))
	for i, p := range resp.GetPayments() {
		payments[i] = paymentToResponse(p)
	}

	writeJSON(w, http.StatusOK, paymentListResponse{
		Payments:   payments,
		TotalCount: resp.GetTotalCount(),
	})
}

// GetPayment returns a single payment with full details.
func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
	paymentID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid payment id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	detail, err := h.paymentClient.GetPayment(ctx, &paymentv1.GetPaymentRequest{
		PaymentId: paymentID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, paymentDetailToResponse(detail))
}

// CreatePayment creates a new payment.
func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req createPaymentRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	payment, err := h.paymentClient.CreatePayment(ctx, &paymentv1.CreatePaymentRequest{
		CustomerId: req.CustomerID,
		StaffId:    req.StaffID,
		RentalId:   req.RentalID,
		Amount:     req.Amount,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, paymentToResponse(payment))
}

// DeletePayment deletes a payment by ID.
func (h *PaymentHandler) DeletePayment(w http.ResponseWriter, r *http.Request) {
	paymentID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid payment id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, err = h.paymentClient.DeletePayment(ctx, &paymentv1.DeletePaymentRequest{
		PaymentId: paymentID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
