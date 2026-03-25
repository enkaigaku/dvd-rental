package model

import "time"

// Rental represents a rental transaction.
type Rental struct {
	RentalID    int32
	RentalDate  time.Time
	InventoryID int32
	CustomerID  int32
	ReturnDate  time.Time // zero value means not yet returned
	StaffID     int32
	LastUpdate  time.Time
}

// RentalDetail is an enriched rental with related entity data.
type RentalDetail struct {
	Rental
	CustomerName string
	FilmTitle    string
	StoreID      int32
}

// ReturnResult is the result of returning a rental, including late fee info.
type ReturnResult struct {
	Rental      Rental
	LateFee     string // "0.00" if on time, e.g. "3.00" for 3 days late
	DaysOverdue int32  // 0 if on time
}

// FilmRentalTerms holds the rental terms for a film associated with an inventory item.
type FilmRentalTerms struct {
	RentalDuration  int16
	RentalRate      string
	ReplacementCost string
	Title           string
	StoreID         int32
}

// Inventory represents a physical DVD copy in a store.
type Inventory struct {
	InventoryID int32
	FilmID      int32
	StoreID     int32
	LastUpdate  time.Time
}
