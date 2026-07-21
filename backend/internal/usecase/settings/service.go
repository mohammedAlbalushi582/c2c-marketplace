// Package settings exposes the admin-editable key/value app settings — the
// listing fee tiers and active-listing duration.
package settings

import (
	"context"
	"strconv"
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

// editable whitelists the keys an admin may change (and how to validate each),
// so the endpoint can't be used to write arbitrary settings.
var editable = map[string]func(string) bool{
	"listing_fee_tier2":      isMoney,
	"listing_fee_tier3_plus": isMoney,
	"listing_duration_days":  isPositiveInt,
}

// All returns every setting as a key→value map.
func (s *Service) All(ctx context.Context) (map[string]string, error) {
	rows, err := s.store.ListSettings(ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(rows))
	for _, r := range rows {
		out[r.Key] = r.Value
	}
	return out, nil
}

// Update validates and upserts the provided settings. Unknown or invalid keys
// are rejected as a whole (no partial writes).
func (s *Service) Update(ctx context.Context, kv map[string]string) error {
	if len(kv) == 0 {
		return domain.ErrInvalidInput
	}
	for k, v := range kv {
		validate, ok := editable[k]
		if !ok || !validate(strings.TrimSpace(v)) {
			return domain.ErrInvalidInput
		}
	}
	for k, v := range kv {
		if err := s.store.UpsertSetting(ctx, sqlc.UpsertSettingParams{Key: k, Value: strings.TrimSpace(v)}); err != nil {
			return err
		}
	}
	return nil
}

func isMoney(s string) bool {
	f, err := strconv.ParseFloat(s, 64)
	return err == nil && f >= 0
}

func isPositiveInt(s string) bool {
	n, err := strconv.Atoi(s)
	return err == nil && n > 0
}
