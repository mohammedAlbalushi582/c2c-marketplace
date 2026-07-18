// Package auth implements registration, login, and token refresh.
package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/alamjad/marketplace/internal/domain"
	pauth "github.com/alamjad/marketplace/internal/platform/auth"
	"github.com/alamjad/marketplace/internal/repository/postgres/sqlc"
)

// Service holds auth business logic.
type Service struct {
	q      sqlc.Querier
	tokens *pauth.Manager
}

func NewService(q sqlc.Querier, tokens *pauth.Manager) *Service {
	return &Service{q: q, tokens: tokens}
}

// RegisterInput carries new-account fields.
type RegisterInput struct {
	Email          string
	Username       *string
	Password       string
	FullName       string
	Phone          *string
	WhatsappNumber *string
}

// Result is the authenticated payload returned to the delivery layer.
type Result struct {
	User         sqlc.User
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64 // access-token lifetime in seconds
}

// Register creates a user and issues an initial token pair.
func (s *Service) Register(ctx context.Context, in RegisterInput) (*Result, error) {
	if _, err := s.q.GetUserByEmail(ctx, in.Email); err == nil {
		return nil, domain.ErrConflict
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	hash, err := pauth.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	user, err := s.q.CreateUser(ctx, sqlc.CreateUserParams{
		Email:          in.Email,
		Username:       in.Username,
		PasswordHash:   &hash,
		FullName:       in.FullName,
		Phone:          in.Phone,
		WhatsappNumber: in.WhatsappNumber,
	})
	if err != nil {
		return nil, err
	}
	return s.issue(ctx, user)
}

// Login authenticates by email or username + password.
func (s *Service) Login(ctx context.Context, identifier, password string) (*Result, error) {
	var (
		user sqlc.User
		err  error
	)
	if strings.Contains(identifier, "@") {
		user, err = s.q.GetUserByEmail(ctx, identifier)
	} else {
		user, err = s.q.GetUserByUsername(ctx, &identifier)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUnauthorized
	} else if err != nil {
		return nil, err
	}

	if user.PasswordHash == nil || !pauth.CheckPassword(*user.PasswordHash, password) {
		return nil, domain.ErrUnauthorized
	}
	if user.Status != sqlc.UserStatusActive {
		return nil, domain.ErrForbidden
	}
	return s.issue(ctx, user)
}

// Refresh rotates a refresh token and issues a new pair.
func (s *Service) Refresh(ctx context.Context, rawRefresh string) (*Result, error) {
	hash := pauth.HashRefreshToken(rawRefresh)
	sess, err := s.q.GetSessionByTokenHash(ctx, hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUnauthorized
	} else if err != nil {
		return nil, err
	}

	user, err := s.q.GetUserByID(ctx, sess.UserID)
	if err != nil {
		return nil, err
	}
	// Rotate: revoke the presented token before issuing a fresh pair.
	if err := s.q.RevokeSession(ctx, hash); err != nil {
		return nil, err
	}
	return s.issue(ctx, user)
}

// Me returns the current user record.
func (s *Service) Me(ctx context.Context, userID int64) (sqlc.User, error) {
	return s.q.GetUserByID(ctx, userID)
}

// issue signs an access token and persists a new refresh session.
func (s *Service) issue(ctx context.Context, user sqlc.User) (*Result, error) {
	access, err := s.tokens.GenerateAccessToken(user.ID, string(user.Role))
	if err != nil {
		return nil, err
	}
	raw, hash, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	if _, err := s.q.CreateSession(ctx, sqlc.CreateSessionParams{
		UserID:           user.ID,
		RefreshTokenHash: hash,
		ExpiresAt:        time.Now().Add(s.tokens.RefreshTTL()),
	}); err != nil {
		return nil, err
	}
	return &Result{
		User:         user,
		AccessToken:  access,
		RefreshToken: raw,
		ExpiresIn:    int64(s.tokens.AccessTTL().Seconds()),
	}, nil
}
