// Package middleware provides HTTP middleware, including JWT authentication.
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/alamjad/marketplace/internal/platform/auth"
)

type ctxKey int

const userCtxKey ctxKey = iota

// AuthUser is the authenticated principal stored in the request context.
type AuthUser struct {
	ID   int64
	Role string
}

// UserLookup re-reads a user's authoritative role and active state. Access
// tokens carry the role they were signed with, so a user demoted or suspended
// mid-session would keep their old powers until the token expires; admin
// routes consult this instead of trusting the claim.
type UserLookup interface {
	LookupUser(ctx context.Context, id int64) (role string, active bool, err error)
}

// Authenticator validates bearer tokens.
type Authenticator struct {
	tokens *auth.Manager
	users  UserLookup
}

func NewAuthenticator(tokens *auth.Manager, users UserLookup) *Authenticator {
	return &Authenticator{tokens: tokens, users: users}
}

// Optional parses a bearer token when present but never rejects the request.
func (a *Authenticator) Optional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if claims, ok := a.parse(r); ok {
			r = r.WithContext(context.WithValue(r.Context(), userCtxKey, AuthUser{ID: claims.UserID, Role: claims.Role}))
		}
		next.ServeHTTP(w, r)
	})
}

// Require rejects requests without a valid bearer token.
func (a *Authenticator) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := a.parse(r)
		if !ok {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userCtxKey, AuthUser{ID: claims.UserID, Role: claims.Role})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAdmin rejects non-admin requests (must run after Require). The role
// is re-read from the database so a revoked admin loses access immediately
// rather than when their access token expires. Fails closed on lookup errors.
func (a *Authenticator) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := UserFromContext(r.Context())
		if !ok {
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
			return
		}
		role, active, err := a.users.LookupUser(r.Context(), u.ID)
		if err != nil || !active || role != "admin" {
			if err != nil {
				slog.Error("admin role lookup failed", "user_id", u.ID, "error", err)
			}
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userCtxKey, AuthUser{ID: u.ID, Role: role})))
	})
}

func (a *Authenticator) parse(r *http.Request) (*auth.Claims, bool) {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return nil, false
	}
	claims, err := a.tokens.ParseAccessToken(strings.TrimPrefix(h, "Bearer "))
	if err != nil {
		return nil, false
	}
	return claims, true
}

// UserFromContext returns the authenticated user, if any.
func UserFromContext(ctx context.Context) (AuthUser, bool) {
	u, ok := ctx.Value(userCtxKey).(AuthUser)
	return u, ok
}
