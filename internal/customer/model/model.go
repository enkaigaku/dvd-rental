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

// Country represents a country.
type Country struct {
	CountryID  int32
	Country    string
	LastUpdate time.Time
}
