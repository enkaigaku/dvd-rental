package handler

import (
	"context"
	"net/http"
	"time"

	paymentv1 "github.com/tokyoyuan/dvd-rental/gen/proto/payment/v1"
	"github.com/tokyoyuan/dvd-rental/pkg/middleware"
)

// PaymentHandler handles payment endpoints (all require auth).
type PaymentHandler struct {
	paymentClient paymentv1.PaymentServiceClient
}

// NewPaymentHandler creates a new PaymentHandler.
func NewPaymentHandler(paymentClient paymentv1.PaymentServiceClient) *PaymentHandler {
	return &PaymentHandler{paymentClient: paymentClient}
}

// --- JSON models ---

type paymentItem struct {
	ID          int32  `json:"id"`
	RentalID    int32  `json:"rental_id"`
	Amount      string `json:"amount"`
	PaymentDate string `json:"payment_date"`
}

type paymentListResponse struct {
	Payments   []paymentItem `json:"payments"`
	TotalCount int32         `json:"total_count"`
	Page       int32         `json:"page"`
	PageSize   int32         `json:"page_size"`
}

// ListPayments returns the authenticated customer's payment history.
func (h *PaymentHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	pageSize, page := parsePagination(r)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.paymentClient.ListPaymentsByCustomer(ctx, &paymentv1.ListPaymentsByCustomerRequest{
		CustomerId: claims.UserID,
		PageSize:   pageSize,
		Page:       page,
	})
	if err != nil {
		grpcToHTTPError(w, err)
		return
	}

	payments := make([]paymentItem, len(resp.GetPayments()))
	for i, p := range resp.GetPayments() {
		payments[i] = paymentItem{
			ID:          p.GetPaymentId(),
			RentalID:    p.GetRentalId(),
			Amount:      p.GetAmount(),
			PaymentDate: timestampToString(p.GetPaymentDate()),
		}
	}

	middleware.WriteJSON(w, http.StatusOK, paymentListResponse{
		Payments:   payments,
		TotalCount: resp.GetTotalCount(),
		Page:       page,
		PageSize:   pageSize,
	})
}
