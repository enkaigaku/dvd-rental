package handler

import (
	"context"

	filmv1 "github.com/tokyoyuan/dvd-rental/gen/proto/film/v1"
	"github.com/tokyoyuan/dvd-rental/internal/film/model"
	"github.com/tokyoyuan/dvd-rental/internal/film/service"
)

// LanguageHandler implements the LanguageService gRPC interface.
type LanguageHandler struct {
	filmv1.UnimplementedLanguageServiceServer
	svc *service.LanguageService
}

// NewLanguageHandler creates a new LanguageHandler.
func NewLanguageHandler(svc *service.LanguageService) *LanguageHandler {
	return &LanguageHandler{svc: svc}
}

func (h *LanguageHandler) GetLanguage(ctx context.Context, req *filmv1.GetLanguageRequest) (*filmv1.Language, error) {
	language, err := h.svc.GetLanguage(ctx, req.GetLanguageId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return languageToProto(language), nil
}

func (h *LanguageHandler) ListLanguages(ctx context.Context, req *filmv1.ListLanguagesRequest) (*filmv1.ListLanguagesResponse, error) {
	languages, total, err := h.svc.ListLanguages(ctx)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toLanguageListResponse(languages, total), nil
}

func toLanguageListResponse(languages []model.Language, total int64) *filmv1.ListLanguagesResponse {
	pbLanguages := make([]*filmv1.Language, len(languages))
	for i, l := range languages {
		pbLanguages[i] = languageToProto(l)
	}
	return &filmv1.ListLanguagesResponse{
		Languages:  pbLanguages,
		TotalCount: int32(total),
	}
}
