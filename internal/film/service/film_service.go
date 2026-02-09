package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/enkaigaku/dvd-rental/internal/film/model"
	"github.com/enkaigaku/dvd-rental/internal/film/repository"
)

var validRatings = map[string]bool{
	"G": true, "PG": true, "PG-13": true, "R": true, "NC-17": true,
}

// FilmService contains business logic for film operations.
type FilmService struct {
	filmRepo     repository.FilmRepository
	actorRepo    repository.ActorRepository
	categoryRepo repository.CategoryRepository
	languageRepo repository.LanguageRepository
}

// NewFilmService creates a new FilmService.
func NewFilmService(
	filmRepo repository.FilmRepository,
	actorRepo repository.ActorRepository,
	categoryRepo repository.CategoryRepository,
	languageRepo repository.LanguageRepository,
) *FilmService {
	return &FilmService{
		filmRepo:     filmRepo,
		actorRepo:    actorRepo,
		categoryRepo: categoryRepo,
		languageRepo: languageRepo,
	}
}

// GetFilm returns a film with enriched details (language names, actors, categories).
func (s *FilmService) GetFilm(ctx context.Context, filmID int32) (model.FilmDetail, error) {
	if filmID <= 0 {
		return model.FilmDetail{}, fmt.Errorf("film_id must be positive: %w", ErrInvalidArgument)
	}

	film, err := s.filmRepo.GetFilm(ctx, filmID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.FilmDetail{}, fmt.Errorf("film %d: %w", filmID, ErrNotFound)
		}
		return model.FilmDetail{}, err
	}

	detail := model.FilmDetail{Film: film}

	// Fetch language name.
	lang, err := s.languageRepo.GetLanguage(ctx, film.LanguageID)
	if err == nil {
		detail.LanguageName = lang.Name
	}

	// Fetch original language name if set.
	if film.OriginalLanguageID != 0 {
		origLang, err := s.languageRepo.GetLanguage(ctx, film.OriginalLanguageID)
		if err == nil {
			detail.OriginalLanguageName = origLang.Name
		}
	}

	// Fetch actors.
	actors, err := s.actorRepo.ListActorsByFilm(ctx, filmID)
	if err == nil {
		detail.Actors = actors
	}

	// Fetch categories.
	categories, err := s.categoryRepo.ListCategoriesByFilm(ctx, filmID)
	if err == nil {
		detail.Categories = categories
	}

	return detail, nil
}

// ListFilms returns a paginated list of films.
func (s *FilmService) ListFilms(ctx context.Context, pageSize, page int32) ([]model.Film, int64, error) {
	pageSize, page = clampPagination(pageSize, page)
	offset := (page - 1) * pageSize

	films, err := s.filmRepo.ListFilms(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.filmRepo.CountFilms(ctx)
	if err != nil {
		return nil, 0, err
	}

	return films, total, nil
}

// SearchFilms performs full-text search on films.
func (s *FilmService) SearchFilms(ctx context.Context, query string, pageSize, page int32) ([]model.Film, int64, error) {
	if query == "" {
		return nil, 0, fmt.Errorf("search query must not be empty: %w", ErrInvalidArgument)
	}

	pageSize, page = clampPagination(pageSize, page)
	offset := (page - 1) * pageSize

	films, err := s.filmRepo.SearchFilms(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.filmRepo.CountSearchFilms(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	return films, total, nil
}

// ListFilmsByCategory returns films in a given category.
func (s *FilmService) ListFilmsByCategory(ctx context.Context, categoryID, pageSize, page int32) ([]model.Film, int64, error) {
	if categoryID <= 0 {
		return nil, 0, fmt.Errorf("category_id must be positive: %w", ErrInvalidArgument)
	}

	pageSize, page = clampPagination(pageSize, page)
	offset := (page - 1) * pageSize

	films, err := s.filmRepo.ListFilmsByCategory(ctx, categoryID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.filmRepo.CountFilmsByCategory(ctx, categoryID)
	if err != nil {
		return nil, 0, err
	}

	return films, total, nil
}

// ListFilmsByActor returns films featuring a given actor.
func (s *FilmService) ListFilmsByActor(ctx context.Context, actorID, pageSize, page int32) ([]model.Film, int64, error) {
	if actorID <= 0 {
		return nil, 0, fmt.Errorf("actor_id must be positive: %w", ErrInvalidArgument)
	}

	pageSize, page = clampPagination(pageSize, page)
	offset := (page - 1) * pageSize

	films, err := s.filmRepo.ListFilmsByActor(ctx, actorID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.filmRepo.CountFilmsByActor(ctx, actorID)
	if err != nil {
		return nil, 0, err
	}

	return films, total, nil
}

// CreateFilm creates a new film after validation.
func (s *FilmService) CreateFilm(ctx context.Context, params repository.CreateFilmParams) (model.Film, error) {
	if err := s.validateFilmParams(ctx, params.Title, params.LanguageID, params.OriginalLanguageID, params.Rating, params.RentalRate, params.ReplacementCost); err != nil {
		return model.Film{}, err
	}

	film, err := s.filmRepo.CreateFilm(ctx, params)
	if err != nil {
		if isForeignKeyViolation(err) {
			return model.Film{}, fmt.Errorf("invalid language_id: %w", ErrInvalidArgument)
		}
		return model.Film{}, err
	}
	return film, nil
}

// UpdateFilm updates an existing film after validation.
func (s *FilmService) UpdateFilm(ctx context.Context, params repository.UpdateFilmParams) (model.Film, error) {
	if params.FilmID <= 0 {
		return model.Film{}, fmt.Errorf("film_id must be positive: %w", ErrInvalidArgument)
	}

	if err := s.validateFilmParams(ctx, params.Title, params.LanguageID, params.OriginalLanguageID, params.Rating, params.RentalRate, params.ReplacementCost); err != nil {
		return model.Film{}, err
	}

	film, err := s.filmRepo.UpdateFilm(ctx, params)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Film{}, fmt.Errorf("film %d: %w", params.FilmID, ErrNotFound)
		}
		if isForeignKeyViolation(err) {
			return model.Film{}, fmt.Errorf("invalid language_id: %w", ErrInvalidArgument)
		}
		return model.Film{}, err
	}
	return film, nil
}

// DeleteFilm deletes a film. Fails if inventory records reference it.
func (s *FilmService) DeleteFilm(ctx context.Context, filmID int32) error {
	if filmID <= 0 {
		return fmt.Errorf("film_id must be positive: %w", ErrInvalidArgument)
	}

	// Verify the film exists.
	_, err := s.filmRepo.GetFilm(ctx, filmID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("film %d: %w", filmID, ErrNotFound)
		}
		return err
	}

	if err := s.filmRepo.DeleteFilm(ctx, filmID); err != nil {
		if isForeignKeyViolation(err) {
			return fmt.Errorf("film %d is referenced by inventory records: %w", filmID, ErrForeignKey)
		}
		return err
	}
	return nil
}

// AddActorToFilm associates an actor with a film.
func (s *FilmService) AddActorToFilm(ctx context.Context, actorID, filmID int32) error {
	if actorID <= 0 {
		return fmt.Errorf("actor_id must be positive: %w", ErrInvalidArgument)
	}
	if filmID <= 0 {
		return fmt.Errorf("film_id must be positive: %w", ErrInvalidArgument)
	}

	// Verify both exist.
	if _, err := s.filmRepo.GetFilm(ctx, filmID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("film %d: %w", filmID, ErrNotFound)
		}
		return err
	}
	if _, err := s.actorRepo.GetActor(ctx, actorID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("actor %d: %w", actorID, ErrNotFound)
		}
		return err
	}

	if err := s.filmRepo.AddActorToFilm(ctx, actorID, filmID); err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("actor %d already associated with film %d: %w", actorID, filmID, ErrAlreadyExists)
		}
		return err
	}
	return nil
}

// RemoveActorFromFilm removes an actor-film association.
func (s *FilmService) RemoveActorFromFilm(ctx context.Context, actorID, filmID int32) error {
	if actorID <= 0 {
		return fmt.Errorf("actor_id must be positive: %w", ErrInvalidArgument)
	}
	if filmID <= 0 {
		return fmt.Errorf("film_id must be positive: %w", ErrInvalidArgument)
	}

	return s.filmRepo.RemoveActorFromFilm(ctx, actorID, filmID)
}

// AddCategoryToFilm associates a category with a film.
func (s *FilmService) AddCategoryToFilm(ctx context.Context, filmID, categoryID int32) error {
	if filmID <= 0 {
		return fmt.Errorf("film_id must be positive: %w", ErrInvalidArgument)
	}
	if categoryID <= 0 {
		return fmt.Errorf("category_id must be positive: %w", ErrInvalidArgument)
	}

	// Verify both exist.
	if _, err := s.filmRepo.GetFilm(ctx, filmID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("film %d: %w", filmID, ErrNotFound)
		}
		return err
	}
	if _, err := s.categoryRepo.GetCategory(ctx, categoryID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("category %d: %w", categoryID, ErrNotFound)
		}
		return err
	}

	if err := s.filmRepo.AddCategoryToFilm(ctx, filmID, categoryID); err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("category %d already associated with film %d: %w", categoryID, filmID, ErrAlreadyExists)
		}
		return err
	}
	return nil
}

// RemoveCategoryFromFilm removes a category-film association.
func (s *FilmService) RemoveCategoryFromFilm(ctx context.Context, filmID, categoryID int32) error {
	if filmID <= 0 {
		return fmt.Errorf("film_id must be positive: %w", ErrInvalidArgument)
	}
	if categoryID <= 0 {
		return fmt.Errorf("category_id must be positive: %w", ErrInvalidArgument)
	}

	return s.filmRepo.RemoveCategoryFromFilm(ctx, filmID, categoryID)
}

// validateFilmParams validates common film creation/update parameters.
func (s *FilmService) validateFilmParams(ctx context.Context, title string, languageID, originalLanguageID int32, rating, rentalRate, replacementCost string) error {
	if title == "" {
		return fmt.Errorf("title must not be empty: %w", ErrInvalidArgument)
	}
	if languageID <= 0 {
		return fmt.Errorf("language_id must be positive: %w", ErrInvalidArgument)
	}

	// Verify language exists.
	if _, err := s.languageRepo.GetLanguage(ctx, languageID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("language %d not found: %w", languageID, ErrInvalidArgument)
		}
		return err
	}

	// Verify original language if provided.
	if originalLanguageID > 0 {
		if _, err := s.languageRepo.GetLanguage(ctx, originalLanguageID); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return fmt.Errorf("original language %d not found: %w", originalLanguageID, ErrInvalidArgument)
			}
			return err
		}
	}

	// Validate rating.
	if rating != "" && !validRatings[rating] {
		return fmt.Errorf("invalid rating %q, must be one of G, PG, PG-13, R, NC-17: %w", rating, ErrInvalidArgument)
	}

	// Validate numeric fields.
	if rentalRate != "" {
		if _, err := strconv.ParseFloat(rentalRate, 64); err != nil {
			return fmt.Errorf("invalid rental_rate %q: %w", rentalRate, ErrInvalidArgument)
		}
	}
	if replacementCost != "" {
		if _, err := strconv.ParseFloat(replacementCost, 64); err != nil {
			return fmt.Errorf("invalid replacement_cost %q: %w", replacementCost, ErrInvalidArgument)
		}
	}

	return nil
}
