package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tokyoyuan/dvd-rental/internal/film/model"
	"github.com/tokyoyuan/dvd-rental/internal/film/repository/sqlcgen"
)

// FilmRepository defines the data access interface for films.
type FilmRepository interface {
	GetFilm(ctx context.Context, filmID int32) (model.Film, error)
	ListFilms(ctx context.Context, limit, offset int32) ([]model.Film, error)
	CountFilms(ctx context.Context) (int64, error)
	SearchFilms(ctx context.Context, query string, limit, offset int32) ([]model.Film, error)
	CountSearchFilms(ctx context.Context, query string) (int64, error)
	ListFilmsByCategory(ctx context.Context, categoryID, limit, offset int32) ([]model.Film, error)
	CountFilmsByCategory(ctx context.Context, categoryID int32) (int64, error)
	ListFilmsByActor(ctx context.Context, actorID, limit, offset int32) ([]model.Film, error)
	CountFilmsByActor(ctx context.Context, actorID int32) (int64, error)
	CreateFilm(ctx context.Context, params CreateFilmParams) (model.Film, error)
	UpdateFilm(ctx context.Context, params UpdateFilmParams) (model.Film, error)
	DeleteFilm(ctx context.Context, filmID int32) error
	AddActorToFilm(ctx context.Context, actorID, filmID int32) error
	RemoveActorFromFilm(ctx context.Context, actorID, filmID int32) error
	AddCategoryToFilm(ctx context.Context, filmID, categoryID int32) error
	RemoveCategoryFromFilm(ctx context.Context, filmID, categoryID int32) error
}

// CreateFilmParams holds parameters for creating a new film.
type CreateFilmParams struct {
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
}

// UpdateFilmParams holds parameters for updating a film.
type UpdateFilmParams struct {
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
}

type filmRepository struct {
	q *sqlcgen.Queries
}

// NewFilmRepository creates a new FilmRepository backed by PostgreSQL.
func NewFilmRepository(pool *pgxpool.Pool) FilmRepository {
	return &filmRepository{q: sqlcgen.New(pool)}
}

func (r *filmRepository) GetFilm(ctx context.Context, filmID int32) (model.Film, error) {
	row, err := r.q.GetFilm(ctx, filmID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Film{}, ErrNotFound
		}
		return model.Film{}, fmt.Errorf("get film: %w", err)
	}
	return filmFromGetRow(row), nil
}

func (r *filmRepository) ListFilms(ctx context.Context, limit, offset int32) ([]model.Film, error) {
	rows, err := r.q.ListFilms(ctx, sqlcgen.ListFilmsParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("list films: %w", err)
	}
	return filmsFromListRows(rows), nil
}

func (r *filmRepository) CountFilms(ctx context.Context) (int64, error) {
	count, err := r.q.CountFilms(ctx)
	if err != nil {
		return 0, fmt.Errorf("count films: %w", err)
	}
	return count, nil
}

func (r *filmRepository) SearchFilms(ctx context.Context, query string, limit, offset int32) ([]model.Film, error) {
	rows, err := r.q.SearchFilms(ctx, sqlcgen.SearchFilmsParams{
		PlaintoTsquery: query,
		Limit:          limit,
		Offset:         offset,
	})
	if err != nil {
		return nil, fmt.Errorf("search films: %w", err)
	}
	return filmsFromSearchRows(rows), nil
}

func (r *filmRepository) CountSearchFilms(ctx context.Context, query string) (int64, error) {
	count, err := r.q.CountSearchFilms(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("count search films: %w", err)
	}
	return count, nil
}

func (r *filmRepository) ListFilmsByCategory(ctx context.Context, categoryID, limit, offset int32) ([]model.Film, error) {
	rows, err := r.q.ListFilmsByCategory(ctx, sqlcgen.ListFilmsByCategoryParams{
		CategoryID: categoryID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list films by category: %w", err)
	}
	return filmsFromCategoryRows(rows), nil
}

func (r *filmRepository) CountFilmsByCategory(ctx context.Context, categoryID int32) (int64, error) {
	count, err := r.q.CountFilmsByCategory(ctx, categoryID)
	if err != nil {
		return 0, fmt.Errorf("count films by category: %w", err)
	}
	return count, nil
}

func (r *filmRepository) ListFilmsByActor(ctx context.Context, actorID, limit, offset int32) ([]model.Film, error) {
	rows, err := r.q.ListFilmsByActor(ctx, sqlcgen.ListFilmsByActorParams{
		ActorID: actorID,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list films by actor: %w", err)
	}
	return filmsFromActorRows(rows), nil
}

func (r *filmRepository) CountFilmsByActor(ctx context.Context, actorID int32) (int64, error) {
	count, err := r.q.CountFilmsByActor(ctx, actorID)
	if err != nil {
		return 0, fmt.Errorf("count films by actor: %w", err)
	}
	return count, nil
}

func (r *filmRepository) CreateFilm(ctx context.Context, params CreateFilmParams) (model.Film, error) {
	row, err := r.q.CreateFilm(ctx, sqlcgen.CreateFilmParams{
		Title:              params.Title,
		Description:        stringToText(params.Description),
		ReleaseYear:        yearFromInt32(params.ReleaseYear),
		LanguageID:         params.LanguageID,
		OriginalLanguageID: int32ToInt4(params.OriginalLanguageID),
		RentalDuration:     params.RentalDuration,
		RentalRate:         stringToNumeric(params.RentalRate),
		Length:             int16ToInt2(params.Length),
		ReplacementCost:    stringToNumeric(params.ReplacementCost),
		Rating:             stringToRating(params.Rating),
		SpecialFeatures:    params.SpecialFeatures,
	})
	if err != nil {
		return model.Film{}, fmt.Errorf("create film: %w", err)
	}
	return filmFromCreateRow(row), nil
}

func (r *filmRepository) UpdateFilm(ctx context.Context, params UpdateFilmParams) (model.Film, error) {
	row, err := r.q.UpdateFilm(ctx, sqlcgen.UpdateFilmParams{
		FilmID:             params.FilmID,
		Title:              params.Title,
		Description:        stringToText(params.Description),
		ReleaseYear:        yearFromInt32(params.ReleaseYear),
		LanguageID:         params.LanguageID,
		OriginalLanguageID: int32ToInt4(params.OriginalLanguageID),
		RentalDuration:     params.RentalDuration,
		RentalRate:         stringToNumeric(params.RentalRate),
		Length:             int16ToInt2(params.Length),
		ReplacementCost:    stringToNumeric(params.ReplacementCost),
		Rating:             stringToRating(params.Rating),
		SpecialFeatures:    params.SpecialFeatures,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Film{}, ErrNotFound
		}
		return model.Film{}, fmt.Errorf("update film: %w", err)
	}
	return filmFromUpdateRow(row), nil
}

func (r *filmRepository) DeleteFilm(ctx context.Context, filmID int32) error {
	if err := r.q.DeleteFilm(ctx, filmID); err != nil {
		return fmt.Errorf("delete film: %w", err)
	}
	return nil
}

func (r *filmRepository) AddActorToFilm(ctx context.Context, actorID, filmID int32) error {
	if err := r.q.AddActorToFilm(ctx, sqlcgen.AddActorToFilmParams{
		ActorID: actorID,
		FilmID:  filmID,
	}); err != nil {
		return fmt.Errorf("add actor to film: %w", err)
	}
	return nil
}

func (r *filmRepository) RemoveActorFromFilm(ctx context.Context, actorID, filmID int32) error {
	if err := r.q.RemoveActorFromFilm(ctx, sqlcgen.RemoveActorFromFilmParams{
		ActorID: actorID,
		FilmID:  filmID,
	}); err != nil {
		return fmt.Errorf("remove actor from film: %w", err)
	}
	return nil
}

func (r *filmRepository) AddCategoryToFilm(ctx context.Context, filmID, categoryID int32) error {
	if err := r.q.AddCategoryToFilm(ctx, sqlcgen.AddCategoryToFilmParams{
		FilmID:     filmID,
		CategoryID: categoryID,
	}); err != nil {
		return fmt.Errorf("add category to film: %w", err)
	}
	return nil
}

func (r *filmRepository) RemoveCategoryFromFilm(ctx context.Context, filmID, categoryID int32) error {
	if err := r.q.RemoveCategoryFromFilm(ctx, sqlcgen.RemoveCategoryFromFilmParams{
		FilmID:     filmID,
		CategoryID: categoryID,
	}); err != nil {
		return fmt.Errorf("remove category from film: %w", err)
	}
	return nil
}

// --- year conversion for INSERT/UPDATE ---

func yearFromInt32(v int32) interface{} {
	if v == 0 {
		return nil
	}
	return v
}

// --- row to model conversions ---

func filmFromGetRow(r sqlcgen.GetFilmRow) model.Film {
	f := convertFilmFields(filmFields{
		FilmID: r.FilmID, Title: r.Title, Description: r.Description,
		ReleaseYear: r.ReleaseYear, LanguageID: r.LanguageID,
		OriginalLanguageID: r.OriginalLanguageID, RentalDuration: r.RentalDuration,
		RentalRate: r.RentalRate, Length: r.Length, ReplacementCost: r.ReplacementCost,
		Rating: r.Rating, SpecialFeatures: r.SpecialFeatures, LastUpdate: r.LastUpdate,
	})
	return filmFromConverted(f)
}

func filmFromCreateRow(r sqlcgen.CreateFilmRow) model.Film {
	f := convertFilmFields(filmFields{
		FilmID: r.FilmID, Title: r.Title, Description: r.Description,
		ReleaseYear: r.ReleaseYear, LanguageID: r.LanguageID,
		OriginalLanguageID: r.OriginalLanguageID, RentalDuration: r.RentalDuration,
		RentalRate: r.RentalRate, Length: r.Length, ReplacementCost: r.ReplacementCost,
		Rating: r.Rating, SpecialFeatures: r.SpecialFeatures, LastUpdate: r.LastUpdate,
	})
	return filmFromConverted(f)
}

func filmFromUpdateRow(r sqlcgen.UpdateFilmRow) model.Film {
	f := convertFilmFields(filmFields{
		FilmID: r.FilmID, Title: r.Title, Description: r.Description,
		ReleaseYear: r.ReleaseYear, LanguageID: r.LanguageID,
		OriginalLanguageID: r.OriginalLanguageID, RentalDuration: r.RentalDuration,
		RentalRate: r.RentalRate, Length: r.Length, ReplacementCost: r.ReplacementCost,
		Rating: r.Rating, SpecialFeatures: r.SpecialFeatures, LastUpdate: r.LastUpdate,
	})
	return filmFromConverted(f)
}

func filmsFromListRows(rows []sqlcgen.ListFilmsRow) []model.Film {
	films := make([]model.Film, len(rows))
	for i, r := range rows {
		f := convertFilmFields(filmFields{
			FilmID: r.FilmID, Title: r.Title, Description: r.Description,
			ReleaseYear: r.ReleaseYear, LanguageID: r.LanguageID,
			OriginalLanguageID: r.OriginalLanguageID, RentalDuration: r.RentalDuration,
			RentalRate: r.RentalRate, Length: r.Length, ReplacementCost: r.ReplacementCost,
			Rating: r.Rating, SpecialFeatures: r.SpecialFeatures, LastUpdate: r.LastUpdate,
		})
		films[i] = filmFromConverted(f)
	}
	return films
}

func filmsFromSearchRows(rows []sqlcgen.SearchFilmsRow) []model.Film {
	films := make([]model.Film, len(rows))
	for i, r := range rows {
		f := convertFilmFields(filmFields{
			FilmID: r.FilmID, Title: r.Title, Description: r.Description,
			ReleaseYear: r.ReleaseYear, LanguageID: r.LanguageID,
			OriginalLanguageID: r.OriginalLanguageID, RentalDuration: r.RentalDuration,
			RentalRate: r.RentalRate, Length: r.Length, ReplacementCost: r.ReplacementCost,
			Rating: r.Rating, SpecialFeatures: r.SpecialFeatures, LastUpdate: r.LastUpdate,
		})
		films[i] = filmFromConverted(f)
	}
	return films
}

func filmsFromCategoryRows(rows []sqlcgen.ListFilmsByCategoryRow) []model.Film {
	films := make([]model.Film, len(rows))
	for i, r := range rows {
		f := convertFilmFields(filmFields{
			FilmID: r.FilmID, Title: r.Title, Description: r.Description,
			ReleaseYear: r.ReleaseYear, LanguageID: r.LanguageID,
			OriginalLanguageID: r.OriginalLanguageID, RentalDuration: r.RentalDuration,
			RentalRate: r.RentalRate, Length: r.Length, ReplacementCost: r.ReplacementCost,
			Rating: r.Rating, SpecialFeatures: r.SpecialFeatures, LastUpdate: r.LastUpdate,
		})
		films[i] = filmFromConverted(f)
	}
	return films
}

func filmsFromActorRows(rows []sqlcgen.ListFilmsByActorRow) []model.Film {
	films := make([]model.Film, len(rows))
	for i, r := range rows {
		f := convertFilmFields(filmFields{
			FilmID: r.FilmID, Title: r.Title, Description: r.Description,
			ReleaseYear: r.ReleaseYear, LanguageID: r.LanguageID,
			OriginalLanguageID: r.OriginalLanguageID, RentalDuration: r.RentalDuration,
			RentalRate: r.RentalRate, Length: r.Length, ReplacementCost: r.ReplacementCost,
			Rating: r.Rating, SpecialFeatures: r.SpecialFeatures, LastUpdate: r.LastUpdate,
		})
		films[i] = filmFromConverted(f)
	}
	return films
}

func filmFromConverted(f filmConvertedFields) model.Film {
	return model.Film{
		FilmID:             f.FilmID,
		Title:              f.Title,
		Description:        f.Description,
		ReleaseYear:        f.ReleaseYear,
		LanguageID:         f.LanguageID,
		OriginalLanguageID: f.OriginalLanguageID,
		RentalDuration:     f.RentalDuration,
		RentalRate:         f.RentalRate,
		Length:             f.Length,
		ReplacementCost:    f.ReplacementCost,
		Rating:             f.Rating,
		SpecialFeatures:    f.SpecialFeatures,
		LastUpdate:         f.LastUpdate,
	}
}
