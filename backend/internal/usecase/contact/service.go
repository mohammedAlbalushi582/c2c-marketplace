// Package contact implements the راسلنا (contact-the-owner) inbox: public
// submissions plus admin read/list/delete.
package contact

import (
	"context"
	"strings"

	"github.com/alamjad/marketplace/internal/domain"
	"github.com/alamjad/marketplace/internal/repository/postgres/sqlc"
)

// Store is the persistence seam (generated queries).
type Store interface {
	sqlc.Querier
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

// CreateInput is a public contact submission.
type CreateInput struct {
	Name    string
	Phone   string
	Email   string
	Message string
}

func (s *Service) Create(ctx context.Context, in CreateInput) (sqlc.ContactMessage, error) {
	name := strings.TrimSpace(in.Name)
	msg := strings.TrimSpace(in.Message)
	if name == "" || msg == "" {
		return sqlc.ContactMessage{}, domain.ErrInvalidInput
	}
	return s.store.CreateContactMessage(ctx, sqlc.CreateContactMessageParams{
		Name:    name,
		Phone:   optional(in.Phone),
		Email:   optional(in.Email),
		Message: msg,
	})
}

// ListResult is a page of messages plus totals for the admin inbox.
type ListResult struct {
	Items  []sqlc.ContactMessage
	Total  int64
	Unread int64
}

func (s *Service) List(ctx context.Context, limit, offset int32) (*ListResult, error) {
	items, err := s.store.ListContactMessages(ctx, sqlc.ListContactMessagesParams{Lim: limit, Off: offset})
	if err != nil {
		return nil, err
	}
	total, err := s.store.CountContactMessages(ctx)
	if err != nil {
		return nil, err
	}
	unread, err := s.store.CountUnreadContactMessages(ctx)
	if err != nil {
		return nil, err
	}
	return &ListResult{Items: items, Total: total, Unread: unread}, nil
}

func (s *Service) MarkRead(ctx context.Context, id int64) error {
	return s.store.MarkContactMessageRead(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.store.DeleteContactMessage(ctx, id)
}

func optional(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}
