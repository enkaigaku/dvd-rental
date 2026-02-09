package handler

import (
	"context"
	"net/http"
	"time"

	filmv1 "github.com/tokyoyuan/dvd-rental/gen/proto/film/v1"
)

// FilmHandler handles film, actor, category, and language management endpoints.
type FilmHandler struct {
	filmClient     filmv1.FilmServiceClient
	actorClient    filmv1.ActorServiceClient
	categoryClient filmv1.CategoryServiceClient
	languageClient filmv1.LanguageServiceClient
}

// NewFilmHandler creates a new FilmHandler.
func NewFilmHandler(
	filmClient filmv1.FilmServiceClient,
	actorClient filmv1.ActorServiceClient,
	categoryClient filmv1.CategoryServiceClient,
	languageClient filmv1.LanguageServiceClient,
) *FilmHandler {
	return &FilmHandler{
		filmClient:     filmClient,
		actorClient:    actorClient,
		categoryClient: categoryClient,
		languageClient: languageClient,
	}
}

// --- JSON models ---

type filmResponse struct {
	FilmID             int32    `json:"film_id"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	ReleaseYear        int32    `json:"release_year"`
	LanguageID         int32    `json:"language_id"`
	OriginalLanguageID int32    `json:"original_language_id,omitempty"`
	RentalDuration     int32    `json:"rental_duration"`
	RentalRate         string   `json:"rental_rate"`
	Length             int32    `json:"length"`
	ReplacementCost    string   `json:"replacement_cost"`
	Rating             string   `json:"rating"`
	SpecialFeatures    []string `json:"special_features"`
	LastUpdate         string   `json:"last_update"`
}

type filmDetailResponse struct {
	filmResponse
	LanguageName         string             `json:"language_name"`
	OriginalLanguageName string             `json:"original_language_name,omitempty"`
	Actors               []actorResponse    `json:"actors"`
	Categories           []categoryResponse `json:"categories"`
}

type filmListResponse struct {
	Films      []filmResponse `json:"films"`
	TotalCount int32          `json:"total_count"`
}

type actorResponse struct {
	ActorID    int32  `json:"actor_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	LastUpdate string `json:"last_update"`
}

type actorListResponse struct {
	Actors     []actorResponse `json:"actors"`
	TotalCount int32           `json:"total_count"`
}

type categoryResponse struct {
	CategoryID int32  `json:"category_id"`
	Name       string `json:"name"`
	LastUpdate string `json:"last_update"`
}

type categoryListResponse struct {
	Categories []categoryResponse `json:"categories"`
	TotalCount int32              `json:"total_count"`
}

type languageResponse struct {
	LanguageID int32  `json:"language_id"`
	Name       string `json:"name"`
	LastUpdate string `json:"last_update"`
}

type languageListResponse struct {
	Languages  []languageResponse `json:"languages"`
	TotalCount int32              `json:"total_count"`
}

type createFilmRequest struct {
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	ReleaseYear        int32    `json:"release_year"`
	LanguageID         int32    `json:"language_id"`
	OriginalLanguageID int32    `json:"original_language_id"`
	RentalDuration     int32    `json:"rental_duration"`
	RentalRate         string   `json:"rental_rate"`
	Length             int32    `json:"length"`
	ReplacementCost    string   `json:"replacement_cost"`
	Rating             string   `json:"rating"`
	SpecialFeatures    []string `json:"special_features"`
}

type updateFilmRequest = createFilmRequest

type createActorRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type updateActorRequest = createActorRequest

type filmActorRequest struct {
	ActorID int32 `json:"actor_id"`
}

type filmCategoryRequest struct {
	CategoryID int32 `json:"category_id"`
}

func filmToResponse(f *filmv1.Film) filmResponse {
	return filmResponse{
		FilmID:             f.GetFilmId(),
		Title:              f.GetTitle(),
		Description:        f.GetDescription(),
		ReleaseYear:        f.GetReleaseYear(),
		LanguageID:         f.GetLanguageId(),
		OriginalLanguageID: f.GetOriginalLanguageId(),
		RentalDuration:     f.GetRentalDuration(),
		RentalRate:         f.GetRentalRate(),
		Length:             f.GetLength(),
		ReplacementCost:    f.GetReplacementCost(),
		Rating:             f.GetRating(),
		SpecialFeatures:    f.GetSpecialFeatures(),
		LastUpdate:         f.GetLastUpdate().AsTime().Format(time.RFC3339),
	}
}

func actorToResponse(a *filmv1.Actor) actorResponse {
	return actorResponse{
		ActorID:    a.GetActorId(),
		FirstName:  a.GetFirstName(),
		LastName:   a.GetLastName(),
		LastUpdate: a.GetLastUpdate().AsTime().Format(time.RFC3339),
	}
}

func categoryToResponse(c *filmv1.Category) categoryResponse {
	return categoryResponse{
		CategoryID: c.GetCategoryId(),
		Name:       c.GetName(),
		LastUpdate: c.GetLastUpdate().AsTime().Format(time.RFC3339),
	}
}

func languageToResponse(l *filmv1.Language) languageResponse {
	return languageResponse{
		LanguageID: l.GetLanguageId(),
		Name:       l.GetName(),
		LastUpdate: l.GetLastUpdate().AsTime().Format(time.RFC3339),
	}
}

// --- Film endpoints ---

// ListFilms returns a paginated list of films.
func (h *FilmHandler) ListFilms(w http.ResponseWriter, r *http.Request) {
	pageSize, page := parsePagination(r)
	categoryID := parseQueryInt32(r, "category_id")
	actorID := parseQueryInt32(r, "actor_id")
	query := r.URL.Query().Get("q")

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var resp *filmv1.ListFilmsResponse
	var err error

	switch {
	case query != "":
		resp, err = h.filmClient.SearchFilms(ctx, &filmv1.SearchFilmsRequest{
			Query:    query,
			PageSize: pageSize,
			Page:     page,
		})
	case categoryID > 0:
		resp, err = h.filmClient.ListFilmsByCategory(ctx, &filmv1.ListFilmsByCategoryRequest{
			CategoryId: categoryID,
			PageSize:   pageSize,
			Page:       page,
		})
	case actorID > 0:
		resp, err = h.filmClient.ListFilmsByActor(ctx, &filmv1.ListFilmsByActorRequest{
			ActorId:  actorID,
			PageSize: pageSize,
			Page:     page,
		})
	default:
		resp, err = h.filmClient.ListFilms(ctx, &filmv1.ListFilmsRequest{
			PageSize: pageSize,
			Page:     page,
		})
	}
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	films := make([]filmResponse, len(resp.GetFilms()))
	for i, f := range resp.GetFilms() {
		films[i] = filmToResponse(f)
	}

	writeJSON(w, http.StatusOK, filmListResponse{
		Films:      films,
		TotalCount: resp.GetTotalCount(),
	})
}

// GetFilm returns a single film with full details.
func (h *FilmHandler) GetFilm(w http.ResponseWriter, r *http.Request) {
	filmID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid film id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	detail, err := h.filmClient.GetFilm(ctx, &filmv1.GetFilmRequest{
		FilmId: filmID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	actors := make([]actorResponse, len(detail.GetActors()))
	for i, a := range detail.GetActors() {
		actors[i] = actorToResponse(a)
	}

	categories := make([]categoryResponse, len(detail.GetCategories()))
	for i, c := range detail.GetCategories() {
		categories[i] = categoryToResponse(c)
	}

	writeJSON(w, http.StatusOK, filmDetailResponse{
		filmResponse:         filmToResponse(detail.GetFilm()),
		LanguageName:         detail.GetLanguageName(),
		OriginalLanguageName: detail.GetOriginalLanguageName(),
		Actors:               actors,
		Categories:           categories,
	})
}

// CreateFilm creates a new film.
func (h *FilmHandler) CreateFilm(w http.ResponseWriter, r *http.Request) {
	var req createFilmRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	film, err := h.filmClient.CreateFilm(ctx, &filmv1.CreateFilmRequest{
		Title:              req.Title,
		Description:        req.Description,
		ReleaseYear:        req.ReleaseYear,
		LanguageId:         req.LanguageID,
		OriginalLanguageId: req.OriginalLanguageID,
		RentalDuration:     req.RentalDuration,
		RentalRate:         req.RentalRate,
		Length:             req.Length,
		ReplacementCost:    req.ReplacementCost,
		Rating:             req.Rating,
		SpecialFeatures:    req.SpecialFeatures,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, filmToResponse(film))
}

// UpdateFilm updates an existing film.
func (h *FilmHandler) UpdateFilm(w http.ResponseWriter, r *http.Request) {
	filmID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid film id")
		return
	}

	var req updateFilmRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	film, err := h.filmClient.UpdateFilm(ctx, &filmv1.UpdateFilmRequest{
		FilmId:             filmID,
		Title:              req.Title,
		Description:        req.Description,
		ReleaseYear:        req.ReleaseYear,
		LanguageId:         req.LanguageID,
		OriginalLanguageId: req.OriginalLanguageID,
		RentalDuration:     req.RentalDuration,
		RentalRate:         req.RentalRate,
		Length:             req.Length,
		ReplacementCost:    req.ReplacementCost,
		Rating:             req.Rating,
		SpecialFeatures:    req.SpecialFeatures,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, filmToResponse(film))
}

// DeleteFilm deletes a film by ID.
func (h *FilmHandler) DeleteFilm(w http.ResponseWriter, r *http.Request) {
	filmID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid film id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, err = h.filmClient.DeleteFilm(ctx, &filmv1.DeleteFilmRequest{
		FilmId: filmID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddActorToFilm adds an actor to a film.
func (h *FilmHandler) AddActorToFilm(w http.ResponseWriter, r *http.Request) {
	filmID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid film id")
		return
	}

	var req filmActorRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, err = h.filmClient.AddActorToFilm(ctx, &filmv1.AddActorToFilmRequest{
		FilmId:  filmID,
		ActorId: req.ActorID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveActorFromFilm removes an actor from a film.
func (h *FilmHandler) RemoveActorFromFilm(w http.ResponseWriter, r *http.Request) {
	filmID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid film id")
		return
	}

	actorID, err := parseIntParam(r, "actorId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid actor id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, err = h.filmClient.RemoveActorFromFilm(ctx, &filmv1.RemoveActorFromFilmRequest{
		FilmId:  filmID,
		ActorId: actorID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddCategoryToFilm adds a category to a film.
func (h *FilmHandler) AddCategoryToFilm(w http.ResponseWriter, r *http.Request) {
	filmID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid film id")
		return
	}

	var req filmCategoryRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, err = h.filmClient.AddCategoryToFilm(ctx, &filmv1.AddCategoryToFilmRequest{
		FilmId:     filmID,
		CategoryId: req.CategoryID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveCategoryFromFilm removes a category from a film.
func (h *FilmHandler) RemoveCategoryFromFilm(w http.ResponseWriter, r *http.Request) {
	filmID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid film id")
		return
	}

	categoryID, err := parseIntParam(r, "categoryId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid category id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, err = h.filmClient.RemoveCategoryFromFilm(ctx, &filmv1.RemoveCategoryFromFilmRequest{
		FilmId:     filmID,
		CategoryId: categoryID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Actor endpoints ---

// ListActors returns a paginated list of actors.
func (h *FilmHandler) ListActors(w http.ResponseWriter, r *http.Request) {
	pageSize, page := parsePagination(r)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.actorClient.ListActors(ctx, &filmv1.ListActorsRequest{
		PageSize: pageSize,
		Page:     page,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	actors := make([]actorResponse, len(resp.GetActors()))
	for i, a := range resp.GetActors() {
		actors[i] = actorToResponse(a)
	}

	writeJSON(w, http.StatusOK, actorListResponse{
		Actors:     actors,
		TotalCount: resp.GetTotalCount(),
	})
}

// GetActor returns a single actor by ID.
func (h *FilmHandler) GetActor(w http.ResponseWriter, r *http.Request) {
	actorID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid actor id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	actor, err := h.actorClient.GetActor(ctx, &filmv1.GetActorRequest{
		ActorId: actorID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, actorToResponse(actor))
}

// CreateActor creates a new actor.
func (h *FilmHandler) CreateActor(w http.ResponseWriter, r *http.Request) {
	var req createActorRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	actor, err := h.actorClient.CreateActor(ctx, &filmv1.CreateActorRequest{
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, actorToResponse(actor))
}

// UpdateActor updates an existing actor.
func (h *FilmHandler) UpdateActor(w http.ResponseWriter, r *http.Request) {
	actorID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid actor id")
		return
	}

	var req updateActorRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	actor, err := h.actorClient.UpdateActor(ctx, &filmv1.UpdateActorRequest{
		ActorId:   actorID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, actorToResponse(actor))
}

// DeleteActor deletes an actor by ID.
func (h *FilmHandler) DeleteActor(w http.ResponseWriter, r *http.Request) {
	actorID, err := parseIntParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid actor id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, err = h.actorClient.DeleteActor(ctx, &filmv1.DeleteActorRequest{
		ActorId: actorID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Category endpoints (read-only) ---

// ListCategories returns all categories.
func (h *FilmHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	pageSize, page := parsePagination(r)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.categoryClient.ListCategories(ctx, &filmv1.ListCategoriesRequest{
		PageSize: pageSize,
		Page:     page,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	categories := make([]categoryResponse, len(resp.GetCategories()))
	for i, c := range resp.GetCategories() {
		categories[i] = categoryToResponse(c)
	}

	writeJSON(w, http.StatusOK, categoryListResponse{
		Categories: categories,
		TotalCount: resp.GetTotalCount(),
	})
}

// --- Language endpoints (read-only) ---

// ListLanguages returns all languages.
func (h *FilmHandler) ListLanguages(w http.ResponseWriter, r *http.Request) {
	pageSize, page := parsePagination(r)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.languageClient.ListLanguages(ctx, &filmv1.ListLanguagesRequest{
		PageSize: pageSize,
		Page:     page,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	languages := make([]languageResponse, len(resp.GetLanguages()))
	for i, l := range resp.GetLanguages() {
		languages[i] = languageToResponse(l)
	}

	writeJSON(w, http.StatusOK, languageListResponse{
		Languages:  languages,
		TotalCount: resp.GetTotalCount(),
	})
}
