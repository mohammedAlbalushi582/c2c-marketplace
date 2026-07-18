// Package middleware provides HTTP middleware, including JWT authentication.
package middleware

import (
	"context"
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

// Authenticator validates bearer tokens.
type Authenticator struct {
	tokens *auth.Manager
}

func NewAuthenticator(tokens *auth.Manager) *Authenticator {
	return &Authenticator{tokens: tokens}
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

// RequireAdmin rejects non-admin requests (must run after Require).
func (a *Authenticator) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := UserFromContext(r.Context())
		if !ok || u.Role != "admin" {
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
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
