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

// Inventory represents a physical DVD copy in a store.
type Inventory struct {
	InventoryID int32
	FilmID      int32
	StoreID     int32
	LastUpdate  time.Time
}
