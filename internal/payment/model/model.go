package model

import "time"

// Payment represents a payment record.
type Payment struct {
	PaymentID   int32
	CustomerID  int32
	StaffID     int32
	RentalID    int32
	Amount      string // numeric(5,2) stored as string
	PaymentDate time.Time
}

// CustomerBalance represents a customer's financial summary.
type CustomerBalance struct {
	CustomerID    int32
	TotalCharges  string
	TotalPayments string
	Balance       string // positive = owes money
	RentalCount   int32
	PaymentCount  int32
}

// StoreRevenue represents revenue for a single store within a date range.
type StoreRevenue struct {
	StoreID      int32
	TotalRevenue string
	PaymentCount int32
	RentalCount  int32
}

// PaymentDetail is an enriched payment with cross-table data.
type PaymentDetail struct {
	Payment
	CustomerName string
	StaffName    string
	RentalDate   time.Time // zero value means rental not found
}
