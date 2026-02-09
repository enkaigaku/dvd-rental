package handler

import (
	"context"
	"net/http"
	"time"

	customerv1 "github.com/tokyoyuan/dvd-rental/gen/proto/customer/v1"
)

// CustomerHandler handles customer management endpoints.
type CustomerHandler struct {
	customerClient customerv1.CustomerServiceClient
}

// NewCustomerHandler creates a new CustomerHandler.
func NewCustomerHandler(customerClient customerv1.CustomerServiceClient) *CustomerHandler {
	return &CustomerHandler{customerClient: customerClient}
}

// --- JSON models ---

type customerResponse struct {
	CustomerID int32  `json:"customer_id"`
	StoreID    int32  `json:"store_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	AddressID  int32  `json:"address_id"`
	Active     bool   `json:"active"`
	CreateDate string `json:"create_date"`
	LastUpdate string `json:"last_update"`
}

type customerDetailResponse struct {
	customerResponse
	Address    string `json:"address"`
	Address2   string `json:"address2"`
	District   string `json:"district"`
	City       string `json:"city"`
	Country    string `json:"country"`
	PostalCode string `json:"postal_code"`
	Phone      string `json:"phone"`
}

type customerListResponse struct {
	Customers  []customerResponse `json:"customers"`
	TotalCount int32              `json:"total_count"`
}

type createCustomerRequest struct {
	StoreID   int32  `json:"store_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	AddressID int32  `json:"address_id"`
	Active    bool   `json:"active"`
}

type updateCustomerRequest struct {
	StoreID   int32  `json:"store_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	AddressID int32  `json:"address_id"`
	Active    bool   `json:"active"`
}

func customerToResponse(c *customerv1.Customer) customerResponse {
	return customerResponse{
		CustomerID: c.GetCustomerId(),
		StoreID:    c.GetStoreId(),
		FirstName:  c.GetFirstName(),
		LastName:   c.GetLastName(),
		Email:      c.GetEmail(),
		AddressID:  c.GetAddressId(),
		Active:     c.GetActive(),
		CreateDate: c.GetCreateDate().AsTime().Format("2006-01-02"),
		LastUpdate: c.GetLastUpdate().AsTime().Format(time.RFC3339),
	}
}

func customerDetailToResponse(d *customerv1.CustomerDetail) customerDetailResponse {
	return customerDetailResponse{
		customerResponse: customerToResponse(d.GetCustomer()),
		Address:          d.GetAddress(),
		Address2:         d.GetAddress2(),
		District:         d.GetDistrict(),
		City:             d.GetCity(),
		Country:          d.GetCountry(),
		PostalCode:       d.GetPostalCode(),
		Phone:            d.GetPhone(),
	}
}

// ListCustomers returns a paginated list of customers.
func (h *CustomerHandler) ListCustomers(w http.ResponseWriter, r *http.Request) {
	pageSize, page := parsePagination(r)
	storeID := parseQueryInt32(r, "store_id")

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var resp *customerv1.ListCustomersResponse
	var err error

	if storeID > 0 {
		resp, err = h.customerClient.ListCustomersByStore(ctx, &customerv1.ListCustomersByStoreRequest{
			StoreId:  storeID,
			PageSize: pageSize,
			Page:     page,
		})
	} else {
		resp, err = h.customerClient.ListCustomers(ctx, &customerv1.ListCustomersRequest{
			PageSize: pageSize,
			Page:     page,
		})
	}
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	customers := make([]customerResponse, len(resp.GetCustomers()))
	for i, c := range resp.GetCustomers() {
		customers[i] = customerToResponse(c)
	}

	writeJSON(w, http.StatusOK, customerListResponse{
		Customers:  customers,
		TotalCount: resp.GetTotalCount(),
	})
}

// GetCustomer returns a single customer with full details.
func (h *CustomerHandler) GetCustomer(w http.ResponseWriter, r *http.Request) {
	customerID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid customer id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	detail, err := h.customerClient.GetCustomer(ctx, &customerv1.GetCustomerRequest{
		CustomerId: customerID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, customerDetailToResponse(detail))
}

// CreateCustomer creates a new customer.
func (h *CustomerHandler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var req createCustomerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	customer, err := h.customerClient.CreateCustomer(ctx, &customerv1.CreateCustomerRequest{
		StoreId:   req.StoreID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		AddressId: req.AddressID,
		Active:    req.Active,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, customerToResponse(customer))
}

// UpdateCustomer updates an existing customer.
func (h *CustomerHandler) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	customerID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid customer id")
		return
	}

	var req updateCustomerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	customer, err := h.customerClient.UpdateCustomer(ctx, &customerv1.UpdateCustomerRequest{
		CustomerId: customerID,
		StoreId:    req.StoreID,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Email:      req.Email,
		AddressId:  req.AddressID,
		Active:     req.Active,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, customerToResponse(customer))
}

// DeleteCustomer deletes a customer by ID.
func (h *CustomerHandler) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	customerID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid customer id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, err = h.customerClient.DeleteCustomer(ctx, &customerv1.DeleteCustomerRequest{
		CustomerId: customerID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
