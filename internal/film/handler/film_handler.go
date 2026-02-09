package handler

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	filmv1 "github.com/enkaigaku/dvd-rental/gen/proto/film/v1"
	"github.com/enkaigaku/dvd-rental/internal/film/model"
	"github.com/enkaigaku/dvd-rental/internal/film/repository"
	"github.com/enkaigaku/dvd-rental/internal/film/service"
)

// FilmHandler implements the FilmService gRPC interface.
type FilmHandler struct {
	filmv1.UnimplementedFilmServiceServer
	svc *service.FilmService
}

// NewFilmHandler creates a new FilmHandler.
func NewFilmHandler(svc *service.FilmService) *FilmHandler {
	return &FilmHandler{svc: svc}
}

func (h *FilmHandler) GetFilm(ctx context.Context, req *filmv1.GetFilmRequest) (*filmv1.FilmDetail, error) {
	detail, err := h.svc.GetFilm(ctx, req.GetFilmId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return filmDetailToProto(detail), nil
}

func (h *FilmHandler) ListFilms(ctx context.Context, req *filmv1.ListFilmsRequest) (*filmv1.ListFilmsResponse, error) {
	films, total, err := h.svc.ListFilms(ctx, req.GetPageSize(), req.GetPage())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toFilmListResponse(films, total), nil
}

func (h *FilmHandler) SearchFilms(ctx context.Context, req *filmv1.SearchFilmsRequest) (*filmv1.ListFilmsResponse, error) {
	films, total, err := h.svc.SearchFilms(ctx, req.GetQuery(), req.GetPageSize(), req.GetPage())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toFilmListResponse(films, total), nil
}

func (h *FilmHandler) ListFilmsByCategory(ctx context.Context, req *filmv1.ListFilmsByCategoryRequest) (*filmv1.ListFilmsResponse, error) {
	films, total, err := h.svc.ListFilmsByCategory(ctx, req.GetCategoryId(), req.GetPageSize(), req.GetPage())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toFilmListResponse(films, total), nil
}

func (h *FilmHandler) ListFilmsByActor(ctx context.Context, req *filmv1.ListFilmsByActorRequest) (*filmv1.ListFilmsResponse, error) {
	films, total, err := h.svc.ListFilmsByActor(ctx, req.GetActorId(), req.GetPageSize(), req.GetPage())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toFilmListResponse(films, total), nil
}

func (h *FilmHandler) CreateFilm(ctx context.Context, req *filmv1.CreateFilmRequest) (*filmv1.Film, error) {
	film, err := h.svc.CreateFilm(ctx, repository.CreateFilmParams{
		Title:              req.GetTitle(),
		Description:        req.GetDescription(),
		ReleaseYear:        req.GetReleaseYear(),
		LanguageID:         req.GetLanguageId(),
		OriginalLanguageID: req.GetOriginalLanguageId(),
		RentalDuration:     int16(req.GetRentalDuration()),
		RentalRate:         req.GetRentalRate(),
		Length:             int16(req.GetLength()),
		ReplacementCost:    req.GetReplacementCost(),
		Rating:             req.GetRating(),
		SpecialFeatures:    req.GetSpecialFeatures(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return filmToProto(film), nil
}

func (h *FilmHandler) UpdateFilm(ctx context.Context, req *filmv1.UpdateFilmRequest) (*filmv1.Film, error) {
	film, err := h.svc.UpdateFilm(ctx, repository.UpdateFilmParams{
		FilmID:             req.GetFilmId(),
		Title:              req.GetTitle(),
		Description:        req.GetDescription(),
		ReleaseYear:        req.GetReleaseYear(),
		LanguageID:         req.GetLanguageId(),
		OriginalLanguageID: req.GetOriginalLanguageId(),
		RentalDuration:     int16(req.GetRentalDuration()),
		RentalRate:         req.GetRentalRate(),
		Length:             int16(req.GetLength()),
		ReplacementCost:    req.GetReplacementCost(),
		Rating:             req.GetRating(),
		SpecialFeatures:    req.GetSpecialFeatures(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return filmToProto(film), nil
}

func (h *FilmHandler) DeleteFilm(ctx context.Context, req *filmv1.DeleteFilmRequest) (*emptypb.Empty, error) {
	if err := h.svc.DeleteFilm(ctx, req.GetFilmId()); err != nil {
		return nil, toGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

func (h *FilmHandler) AddActorToFilm(ctx context.Context, req *filmv1.AddActorToFilmRequest) (*emptypb.Empty, error) {
	if err := h.svc.AddActorToFilm(ctx, req.GetActorId(), req.GetFilmId()); err != nil {
		return nil, toGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

func (h *FilmHandler) RemoveActorFromFilm(ctx context.Context, req *filmv1.RemoveActorFromFilmRequest) (*emptypb.Empty, error) {
	if err := h.svc.RemoveActorFromFilm(ctx, req.GetActorId(), req.GetFilmId()); err != nil {
		return nil, toGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

func (h *FilmHandler) AddCategoryToFilm(ctx context.Context, req *filmv1.AddCategoryToFilmRequest) (*emptypb.Empty, error) {
	if err := h.svc.AddCategoryToFilm(ctx, req.GetFilmId(), req.GetCategoryId()); err != nil {
		return nil, toGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

func (h *FilmHandler) RemoveCategoryFromFilm(ctx context.Context, req *filmv1.RemoveCategoryFromFilmRequest) (*emptypb.Empty, error) {
	if err := h.svc.RemoveCategoryFromFilm(ctx, req.GetFilmId(), req.GetCategoryId()); err != nil {
		return nil, toGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

// --- helpers ---

func toFilmListResponse(films []model.Film, total int64) *filmv1.ListFilmsResponse {
	pbFilms := make([]*filmv1.Film, len(films))
	for i, f := range films {
		pbFilms[i] = filmToProto(f)
	}
	return &filmv1.ListFilmsResponse{
		Films:      pbFilms,
		TotalCount: int32(total),
	}
}
