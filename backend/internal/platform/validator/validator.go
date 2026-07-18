// Package validator wraps go-playground/validator for struct validation.
package validator

import "github.com/go-playground/validator/v10"

var v = validator.New()

// Struct validates a struct by its `validate` tags.
func Struct(s any) error { return v.Struct(s) }
