package model

import "time"

// Customer represents a customer in the DVD rental system.
type Customer struct {
	CustomerID   int32
	StoreID      int32
	FirstName    string
	LastName     string
	Email        string
	AddressID    int32
	Active       bool
	CreateDate   time.Time
	LastUpdate   time.Time
	PasswordHash string // Only populated by GetCustomerByEmail for BFF auth.
}

// CustomerDetail is an enriched customer with address information.
type CustomerDetail struct {
	Customer
	Address     string
	Address2    string
	District    string
	CityName    string
	CountryName string
	PostalCode  string
	Phone       string
}

// Address represents a physical address.
type Address struct {
	AddressID  int32
	Address    string
	Address2   string
	District   string
	CityID     int32
	PostalCode string
	Phone      string
	LastUpdate time.Time
}

// City represents a city.
type City struct {
	CityID     int32
	City       string
	CountryID  int32
	LastUpdate time.Time
}

// CustomerStanding represents the account standing of a customer.
type CustomerStanding struct {
	CustomerID         int32
	InGoodStanding     bool
	Reasons            []string
	ActiveRentals      int32
	OverdueRentals     int32
	OutstandingBalance string
}

// CustomerSummary is an aggregate view of a customer's rental activity.
type CustomerSummary struct {
	CustomerID         int32
	TotalRentals       int32
	ActiveRentals      int32
	TotalSpent         string
	FavoriteCategory   string
	OutstandingBalance string
}

// Country represents a country.
type Country struct {
	CountryID  int32
	Country    string
	LastUpdate time.Time
}
