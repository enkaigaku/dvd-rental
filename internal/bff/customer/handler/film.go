package handler

import (
	"context"
	"net/http"
	"time"

	filmv1 "github.com/enkaigaku/dvd-rental/gen/proto/film/v1"
	"github.com/enkaigaku/dvd-rental/pkg/middleware"
)

// FilmHandler handles film catalog endpoints (all public, read-only).
type FilmHandler struct {
	filmClient     filmv1.FilmServiceClient
	actorClient    filmv1.ActorServiceClient
	categoryClient filmv1.CategoryServiceClient
}

// NewFilmHandler creates a new FilmHandler.
func NewFilmHandler(
	filmClient filmv1.FilmServiceClient,
	actorClient filmv1.ActorServiceClient,
	categoryClient filmv1.CategoryServiceClient,
) *FilmHandler {
	return &FilmHandler{
		filmClient:     filmClient,
		actorClient:    actorClient,
		categoryClient: categoryClient,
	}
}

// --- JSON models ---

type filmListItem struct {
	ID          int32  `json:"id"`
	Title       string `json:"title"`
	ReleaseYear int32  `json:"release_year,omitempty"`
	RentalRate  string `json:"rental_rate"`
	Length      int32  `json:"length,omitempty"`
	Rating      string `json:"rating,omitempty"`
}

type filmDetailResponse struct {
	ID               int32          `json:"id"`
	Title            string         `json:"title"`
	Description      string         `json:"description,omitempty"`
	ReleaseYear      int32          `json:"release_year,omitempty"`
	RentalRate       string         `json:"rental_rate"`
	RentalDuration   int32          `json:"rental_duration"`
	Length           int32          `json:"length,omitempty"`
	ReplacementCost  string         `json:"replacement_cost"`
	Rating           string         `json:"rating,omitempty"`
	SpecialFeatures  []string       `json:"special_features,omitempty"`
	Language         string         `json:"language,omitempty"`
	OriginalLanguage string         `json:"original_language,omitempty"`
	Actors           []actorItem    `json:"actors,omitempty"`
	Categories       []categoryItem `json:"categories,omitempty"`
}

type actorItem struct {
	ID        int32  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type categoryItem struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type filmListResponse struct {
	Films      []filmListItem `json:"films"`
	TotalCount int32          `json:"total_count"`
	Page       int32          `json:"page"`
	PageSize   int32          `json:"page_size"`
}

// filmsToListItems converts proto films to JSON list items.
func filmsToListItems(films []*filmv1.Film) []filmListItem {
	items := make([]filmListItem, len(films))
	for i, f := range films {
		items[i] = filmListItem{
			ID:          f.GetFilmId(),
			Title:       f.GetTitle(),
			ReleaseYear: f.GetReleaseYear(),
			RentalRate:  f.GetRentalRate(),
			Length:      f.GetLength(),
			Rating:      f.GetRating(),
		}
	}
	return items
}

// ListFilms returns a paginated film list.
func (h *FilmHandler) ListFilms(w http.ResponseWriter, r *http.Request) {
	pageSize, page := parsePagination(r)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.filmClient.ListFilms(ctx, &filmv1.ListFilmsRequest{
		PageSize: pageSize,
		Page:     page,
	})
	if err != nil {
		grpcToHTTPError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, filmListResponse{
		Films:      filmsToListItems(resp.GetFilms()),
		TotalCount: resp.GetTotalCount(),
		Page:       page,
		PageSize:   pageSize,
	})
}

// GetFilm returns detailed film information.
func (h *FilmHandler) GetFilm(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	detail, err := h.filmClient.GetFilm(ctx, &filmv1.GetFilmRequest{FilmId: id})
	if err != nil {
		grpcToHTTPError(w, err)
		return
	}

	f := detail.GetFilm()

	actors := make([]actorItem, len(detail.GetActors()))
	for i, a := range detail.GetActors() {
		actors[i] = actorItem{
			ID:        a.GetActorId(),
			FirstName: a.GetFirstName(),
			LastName:  a.GetLastName(),
		}
	}

	categories := make([]categoryItem, len(detail.GetCategories()))
	for i, c := range detail.GetCategories() {
		categories[i] = categoryItem{
			ID:   c.GetCategoryId(),
			Name: c.GetName(),
		}
	}

	middleware.WriteJSON(w, http.StatusOK, filmDetailResponse{
		ID:               f.GetFilmId(),
		Title:            f.GetTitle(),
		Description:      f.GetDescription(),
		ReleaseYear:      f.GetReleaseYear(),
		RentalRate:       f.GetRentalRate(),
		RentalDuration:   f.GetRentalDuration(),
		Length:           f.GetLength(),
		ReplacementCost:  f.GetReplacementCost(),
		Rating:           f.GetRating(),
		SpecialFeatures:  f.GetSpecialFeatures(),
		Language:         detail.GetLanguageName(),
		OriginalLanguage: detail.GetOriginalLanguageName(),
		Actors:           actors,
		Categories:       categories,
	})
}

// SearchFilms searches films by query string.
func (h *FilmHandler) SearchFilms(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "query parameter 'q' is required")
		return
	}
	pageSize, page := parsePagination(r)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.filmClient.SearchFilms(ctx, &filmv1.SearchFilmsRequest{
		Query:    q,
		PageSize: pageSize,
		Page:     page,
	})
	if err != nil {
		grpcToHTTPError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, filmListResponse{
		Films:      filmsToListItems(resp.GetFilms()),
		TotalCount: resp.GetTotalCount(),
		Page:       page,
		PageSize:   pageSize,
	})
}

// ListFilmsByCategory returns films in a specific category.
func (h *FilmHandler) ListFilmsByCategory(w http.ResponseWriter, r *http.Request) {
	categoryID, err := parseID(r, "id")
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	pageSize, page := parsePagination(r)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.filmClient.ListFilmsByCategory(ctx, &filmv1.ListFilmsByCategoryRequest{
		CategoryId: categoryID,
		PageSize:   pageSize,
		Page:       page,
	})
	if err != nil {
		grpcToHTTPError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, filmListResponse{
		Films:      filmsToListItems(resp.GetFilms()),
		TotalCount: resp.GetTotalCount(),
		Page:       page,
		PageSize:   pageSize,
	})
}

// ListFilmsByActor returns films featuring a specific actor.
func (h *FilmHandler) ListFilmsByActor(w http.ResponseWriter, r *http.Request) {
	actorID, err := parseID(r, "id")
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	pageSize, page := parsePagination(r)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.filmClient.ListFilmsByActor(ctx, &filmv1.ListFilmsByActorRequest{
		ActorId:  actorID,
		PageSize: pageSize,
		Page:     page,
	})
	if err != nil {
		grpcToHTTPError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, filmListResponse{
		Films:      filmsToListItems(resp.GetFilms()),
		TotalCount: resp.GetTotalCount(),
		Page:       page,
		PageSize:   pageSize,
	})
}

// ListCategories returns all film categories.
func (h *FilmHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	pageSize, page := parsePagination(r)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.categoryClient.ListCategories(ctx, &filmv1.ListCategoriesRequest{
		PageSize: pageSize,
		Page:     page,
	})
	if err != nil {
		grpcToHTTPError(w, err)
		return
	}

	categories := make([]categoryItem, len(resp.GetCategories()))
	for i, c := range resp.GetCategories() {
		categories[i] = categoryItem{
			ID:   c.GetCategoryId(),
			Name: c.GetName(),
		}
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"categories":  categories,
		"total_count": resp.GetTotalCount(),
		"page":        page,
		"page_size":   pageSize,
	})
}

// ListActors returns a paginated actor list.
func (h *FilmHandler) ListActors(w http.ResponseWriter, r *http.Request) {
	pageSize, page := parsePagination(r)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.actorClient.ListActors(ctx, &filmv1.ListActorsRequest{
		PageSize: pageSize,
		Page:     page,
	})
	if err != nil {
		grpcToHTTPError(w, err)
		return
	}

	actors := make([]actorItem, len(resp.GetActors()))
	for i, a := range resp.GetActors() {
		actors[i] = actorItem{
			ID:        a.GetActorId(),
			FirstName: a.GetFirstName(),
			LastName:  a.GetLastName(),
		}
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"actors":      actors,
		"total_count": resp.GetTotalCount(),
		"page":        page,
		"page_size":   pageSize,
	})
}
