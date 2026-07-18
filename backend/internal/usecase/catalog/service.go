// Package catalog implements category, dynamic-field, and location logic.
package catalog

import (
	"context"

	"github.com/alamjad/marketplace/internal/repository/postgres/sqlc"
)

type Service struct {
	q sqlc.Querier
}

func NewService(q sqlc.Querier) *Service {
	return &Service{q: q}
}

// ---- reads ----

func (s *Service) ListCategories(ctx context.Context) ([]sqlc.Category, error) {
	return s.q.ListCategories(ctx)
}

func (s *Service) GetCategoryByID(ctx context.Context, id int64) (sqlc.Category, error) {
	return s.q.GetCategoryByID(ctx, id)
}

func (s *Service) GetCategoryBySlug(ctx context.Context, slug string) (sqlc.Category, error) {
	return s.q.GetCategoryBySlug(ctx, slug)
}

// ResolvedFields returns the category's own fields plus those inherited from ancestors.
func (s *Service) ResolvedFields(ctx context.Context, categoryID int64) ([]sqlc.CategoryField, error) {
	return s.q.ListResolvedFieldsForCategory(ctx, categoryID)
}

func (s *Service) ListGovernorates(ctx context.Context) ([]sqlc.Location, error) {
	return s.q.ListGovernorates(ctx)
}

func (s *Service) ListWilayats(ctx context.Context, governorateID int64) ([]sqlc.Location, error) {
	return s.q.ListWilayatsByGovernorate(ctx, &governorateID)
}

func (s *Service) ListAllLocations(ctx context.Context) ([]sqlc.Location, error) {
	return s.q.ListLocations(ctx)
}

// ---- admin writes ----

func (s *Service) CreateCategory(ctx context.Context, p sqlc.CreateCategoryParams) (sqlc.Category, error) {
	return s.q.CreateCategory(ctx, p)
}

func (s *Service) UpdateCategory(ctx context.Context, p sqlc.UpdateCategoryParams) (sqlc.Category, error) {
	return s.q.UpdateCategory(ctx, p)
}

func (s *Service) DeleteCategory(ctx context.Context, id int64) error {
	return s.q.DeleteCategory(ctx, id)
}

func (s *Service) CreateField(ctx context.Context, p sqlc.CreateCategoryFieldParams) (sqlc.CategoryField, error) {
	return s.q.CreateCategoryField(ctx, p)
}

func (s *Service) DeleteField(ctx context.Context, id int64) error {
	return s.q.DeleteCategoryField(ctx, id)
}
