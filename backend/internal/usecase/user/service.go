// Package user implements admin user management: listing accounts, changing
// roles, and suspending or reactivating users.
package user

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/alamjad/marketplace/internal/domain"
	"github.com/alamjad/marketplace/internal/repository/postgres/sqlc"
)

type Service struct {
	q sqlc.Querier
}

func NewService(q sqlc.Querier) *Service {
	return &Service{q: q}
}

// LookupUser returns a user's current role and whether their account is
// active. It backs the admin middleware's per-request authorization check.
func (s *Service) LookupUser(ctx context.Context, id int64) (string, bool, error) {
	u, err := s.q.GetUserByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	} else if err != nil {
		return "", false, err
	}
	return string(u.Role), u.Status == sqlc.UserStatusActive, nil
}

// ListResult is a page of users plus the unpaginated total.
type ListResult struct {
	Items []sqlc.AdminListUsersRow
	Total int64
}

// List returns users for the admin panel, optionally filtered by role and a
// keyword matched against email, full name, and username.
func (s *Service) List(ctx context.Context, keyword, role string, limit, offset int32) (*ListResult, error) {
	var nr sqlc.NullUserRole
	if role != "" {
		r := sqlc.UserRole(role)
		if !r.Valid() {
			return nil, domain.ErrInvalidInput
		}
		nr = sqlc.NullUserRole{UserRole: r, Valid: true}
	}
	var kw *string
	if keyword != "" {
		kw = &keyword
	}

	items, err := s.q.AdminListUsers(ctx, sqlc.AdminListUsersParams{
		Role: nr, Keyword: kw, Lim: limit, Off: offset,
	})
	if err != nil {
		return nil, err
	}
	total, err := s.q.CountUsers(ctx, sqlc.CountUsersParams{Role: nr, Keyword: kw})
	if err != nil {
		return nil, err
	}
	return &ListResult{Items: items, Total: total}, nil
}

// SetRole promotes or demotes a user. Admins may not change their own role,
// and the last active admin may not be demoted — either would risk locking
// everyone out of the admin panel.
func (s *Service) SetRole(ctx context.Context, actorID, targetID int64, role string) (sqlc.User, error) {
	if actorID == targetID {
		return sqlc.User{}, domain.ErrForbidden
	}
	r := sqlc.UserRole(role)
	if !r.Valid() {
		return sqlc.User{}, domain.ErrInvalidInput
	}

	target, err := s.q.GetUserByID(ctx, targetID)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.User{}, domain.ErrNotFound
	} else if err != nil {
		return sqlc.User{}, err
	}

	// Only an *active* admin counts toward the lockout risk; demoting an
	// already-suspended admin cannot lock anyone out.
	if target.Role == sqlc.UserRoleAdmin && target.Status == sqlc.UserStatusActive && r != sqlc.UserRoleAdmin {
		if err := s.assertNotLastAdmin(ctx); err != nil {
			return sqlc.User{}, err
		}
	}
	u, err := s.q.AdminSetUserRole(ctx, sqlc.AdminSetUserRoleParams{ID: targetID, Role: r})
	if err != nil {
		return sqlc.User{}, err
	}
	// Force a re-login so the new role is reflected in freshly signed tokens.
	if err := s.q.RevokeAllUserSessions(ctx, targetID); err != nil {
		return sqlc.User{}, err
	}
	return u, nil
}

// SetStatus suspends, reactivates, or soft-deletes a user. Suspending or
// deleting also revokes every refresh token they hold, so an already-issued
// session cannot outlive the ban. Admins may not suspend themselves, and the
// last active admin is protected.
func (s *Service) SetStatus(ctx context.Context, actorID, targetID int64, status string) (sqlc.User, error) {
	if actorID == targetID {
		return sqlc.User{}, domain.ErrForbidden
	}
	st := sqlc.UserStatus(status)
	if !st.Valid() {
		return sqlc.User{}, domain.ErrInvalidInput
	}

	target, err := s.q.GetUserByID(ctx, targetID)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.User{}, domain.ErrNotFound
	} else if err != nil {
		return sqlc.User{}, err
	}

	if target.Role == sqlc.UserRoleAdmin && st != sqlc.UserStatusActive {
		if err := s.assertNotLastAdmin(ctx); err != nil {
			return sqlc.User{}, err
		}
	}

	u, err := s.q.AdminSetUserStatus(ctx, sqlc.AdminSetUserStatusParams{ID: targetID, Status: st})
	if err != nil {
		return sqlc.User{}, err
	}
	if st != sqlc.UserStatusActive {
		if err := s.q.RevokeAllUserSessions(ctx, targetID); err != nil {
			return sqlc.User{}, err
		}
	}
	return u, nil
}

// assertNotLastAdmin fails when only one active admin remains.
func (s *Service) assertNotLastAdmin(ctx context.Context) error {
	n, err := s.q.CountAdmins(ctx)
	if err != nil {
		return err
	}
	if n <= 1 {
		return domain.ErrForbidden
	}
	return nil
}
