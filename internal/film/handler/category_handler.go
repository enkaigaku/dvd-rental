package handler

import (
	"context"

	filmv1 "github.com/tokyoyuan/dvd-rental/gen/proto/film/v1"
	"github.com/tokyoyuan/dvd-rental/internal/film/model"
	"github.com/tokyoyuan/dvd-rental/internal/film/service"
)

// CategoryHandler implements the CategoryService gRPC interface.
type CategoryHandler struct {
	filmv1.UnimplementedCategoryServiceServer
	svc *service.CategoryService
}

// NewCategoryHandler creates a new CategoryHandler.
func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

func (h *CategoryHandler) GetCategory(ctx context.Context, req *filmv1.GetCategoryRequest) (*filmv1.Category, error) {
	category, err := h.svc.GetCategory(ctx, req.GetCategoryId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return categoryToProto(category), nil
}

func (h *CategoryHandler) ListCategories(ctx context.Context, req *filmv1.ListCategoriesRequest) (*filmv1.ListCategoriesResponse, error) {
	categories, total, err := h.svc.ListCategories(ctx)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toCategoryListResponse(categories, total), nil
}

func toCategoryListResponse(categories []model.Category, total int64) *filmv1.ListCategoriesResponse {
	pbCategories := make([]*filmv1.Category, len(categories))
	for i, c := range categories {
		pbCategories[i] = categoryToProto(c)
	}
	return &filmv1.ListCategoriesResponse{
		Categories: pbCategories,
		TotalCount: int32(total),
	}
}
