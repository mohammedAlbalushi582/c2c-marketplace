// Package domain holds core entities and repository interfaces (ports).
// It has no dependencies on the database, HTTP, or any framework.
package domain

import "errors"

// Sentinel errors mapped to HTTP status codes by the delivery layer.
var (
	ErrNotFound     = errors.New("resource not found")
	ErrConflict     = errors.New("resource already exists")
	ErrInvalidInput = errors.New("invalid input")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)
