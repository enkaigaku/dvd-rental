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

// PaymentDetail is an enriched payment with cross-table data.
type PaymentDetail struct {
	Payment
	CustomerName string
	StaffName    string
	RentalDate   time.Time // zero value means rental not found
}
