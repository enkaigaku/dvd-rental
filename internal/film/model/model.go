package model

import "time"

// Film represents a film in the DVD rental system.
type Film struct {
	FilmID             int32
	Title              string
	Description        string
	ReleaseYear        int32
	LanguageID         int32
	OriginalLanguageID int32
	RentalDuration     int16
	RentalRate         string
	Length             int16
	ReplacementCost    string
	Rating             string
	SpecialFeatures    []string
	LastUpdate         time.Time
}

// FilmDetail is an enriched film with related entity data, used for single-film views.
type FilmDetail struct {
	Film
	LanguageName         string
	OriginalLanguageName string
	Actors               []Actor
	Categories           []Category
}

// Actor represents a film actor.
type Actor struct {
	ActorID    int32
	FirstName  string
	LastName   string
	LastUpdate time.Time
}

// Category represents a film category.
type Category struct {
	CategoryID int32
	Name       string
	LastUpdate time.Time
}

// Language represents a language used for films.
type Language struct {
	LanguageID int32
	Name       string
	LastUpdate time.Time
}
