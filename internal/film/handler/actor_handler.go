package handler

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	filmv1 "github.com/enkaigaku/dvd-rental/gen/proto/film/v1"
	"github.com/enkaigaku/dvd-rental/internal/film/model"
	"github.com/enkaigaku/dvd-rental/internal/film/service"
)

// ActorHandler implements the ActorService gRPC interface.
type ActorHandler struct {
	filmv1.UnimplementedActorServiceServer
	svc *service.ActorService
}

// NewActorHandler creates a new ActorHandler.
func NewActorHandler(svc *service.ActorService) *ActorHandler {
	return &ActorHandler{svc: svc}
}

func (h *ActorHandler) GetActor(ctx context.Context, req *filmv1.GetActorRequest) (*filmv1.Actor, error) {
	actor, err := h.svc.GetActor(ctx, req.GetActorId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return actorToProto(actor), nil
}

func (h *ActorHandler) ListActors(ctx context.Context, req *filmv1.ListActorsRequest) (*filmv1.ListActorsResponse, error) {
	actors, total, err := h.svc.ListActors(ctx, req.GetPageSize(), req.GetPage())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toActorListResponse(actors, total), nil
}

func (h *ActorHandler) ListActorsByFilm(ctx context.Context, req *filmv1.ListActorsByFilmRequest) (*filmv1.ListActorsResponse, error) {
	actors, err := h.svc.ListActorsByFilm(ctx, req.GetFilmId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toActorListResponse(actors, int64(len(actors))), nil
}

func (h *ActorHandler) CreateActor(ctx context.Context, req *filmv1.CreateActorRequest) (*filmv1.Actor, error) {
	actor, err := h.svc.CreateActor(ctx, req.GetFirstName(), req.GetLastName())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return actorToProto(actor), nil
}

func (h *ActorHandler) UpdateActor(ctx context.Context, req *filmv1.UpdateActorRequest) (*filmv1.Actor, error) {
	actor, err := h.svc.UpdateActor(ctx, req.GetActorId(), req.GetFirstName(), req.GetLastName())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return actorToProto(actor), nil
}

func (h *ActorHandler) DeleteActor(ctx context.Context, req *filmv1.DeleteActorRequest) (*emptypb.Empty, error) {
	if err := h.svc.DeleteActor(ctx, req.GetActorId()); err != nil {
		return nil, toGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

func toActorListResponse(actors []model.Actor, total int64) *filmv1.ListActorsResponse {
	pbActors := make([]*filmv1.Actor, len(actors))
	for i, a := range actors {
		pbActors[i] = actorToProto(a)
	}
	return &filmv1.ListActorsResponse{
		Actors:     pbActors,
		TotalCount: int32(total),
	}
}
