// Package model defines the domain models for the store service.
package model

import "time"

// Store represents a DVD rental store.
type Store struct {
	StoreID        int32
	ManagerStaffID int32
	AddressID      int32
	LastUpdate     time.Time
}

// Staff represents a staff member working at a store.
type Staff struct {
	StaffID      int32
	FirstName    string
	LastName     string
	AddressID    int32
	Email        string
	StoreID      int32
	Active       bool
	Username     string
	PasswordHash string
	Picture      []byte
	LastUpdate   time.Time
}
