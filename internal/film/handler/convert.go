package handler

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	filmv1 "github.com/enkaigaku/dvd-rental/gen/proto/film/v1"
	"github.com/enkaigaku/dvd-rental/internal/film/model"
	"github.com/enkaigaku/dvd-rental/internal/film/service"
)

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, service.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, service.ErrForeignKey):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func filmToProto(f model.Film) *filmv1.Film {
	return &filmv1.Film{
		FilmId:             f.FilmID,
		Title:              f.Title,
		Description:        f.Description,
		ReleaseYear:        f.ReleaseYear,
		LanguageId:         f.LanguageID,
		OriginalLanguageId: f.OriginalLanguageID,
		RentalDuration:     int32(f.RentalDuration),
		RentalRate:         f.RentalRate,
		Length:             int32(f.Length),
		ReplacementCost:    f.ReplacementCost,
		Rating:             f.Rating,
		SpecialFeatures:    f.SpecialFeatures,
		LastUpdate:         timestamppb.New(f.LastUpdate),
	}
}

func filmDetailToProto(d model.FilmDetail) *filmv1.FilmDetail {
	actors := make([]*filmv1.Actor, len(d.Actors))
	for i, a := range d.Actors {
		actors[i] = actorToProto(a)
	}

	categories := make([]*filmv1.Category, len(d.Categories))
	for i, c := range d.Categories {
		categories[i] = categoryToProto(c)
	}

	return &filmv1.FilmDetail{
		Film:                 filmToProto(d.Film),
		LanguageName:         d.LanguageName,
		OriginalLanguageName: d.OriginalLanguageName,
		Actors:               actors,
		Categories:           categories,
	}
}

func actorToProto(a model.Actor) *filmv1.Actor {
	return &filmv1.Actor{
		ActorId:    a.ActorID,
		FirstName:  a.FirstName,
		LastName:   a.LastName,
		LastUpdate: timestamppb.New(a.LastUpdate),
	}
}

func categoryToProto(c model.Category) *filmv1.Category {
	return &filmv1.Category{
		CategoryId: c.CategoryID,
		Name:       c.Name,
		LastUpdate: timestamppb.New(c.LastUpdate),
	}
}

func languageToProto(l model.Language) *filmv1.Language {
	return &filmv1.Language{
		LanguageId: l.LanguageID,
		Name:       l.Name,
		LastUpdate: timestamppb.New(l.LastUpdate),
	}
}
