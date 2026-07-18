// Package listing implements listing CRUD, search, images, EAV attributes,
// favorites, and admin moderation.
package listing

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"unicode"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/alamjad/marketplace/internal/domain"
	"github.com/alamjad/marketplace/internal/platform/pgutil"
	"github.com/alamjad/marketplace/internal/repository/postgres/sqlc"
)

// Store is the persistence seam: generated queries plus transaction support.
type Store interface {
	sqlc.Querier
	WithTx(ctx context.Context, fn func(*sqlc.Queries) error) error
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

// AttributeInput is a single category-field value supplied by the client.
type AttributeInput struct {
	FieldID int64
	Value   any
}

// CreateInput carries the fields needed to create a listing.
type CreateInput struct {
	UserID         int64
	CategoryID     int64
	LocationID     *int64
	Title          string
	Description    string
	Price          *float64
	PriceType      string
	ContactPhone   *string
	WhatsappNumber *string
	Attributes     []AttributeInput
}

// Detail is the full listing view returned for detail pages.
type Detail struct {
	Listing    sqlc.GetListingByIDRow
	Images     []sqlc.ListingImage
	Attributes []sqlc.ListAttributesByListingRow
}

// SearchResult bundles a page of results with the total count.
type SearchResult struct {
	Items []sqlc.SearchListingsRow
	Total int64
}

// Create inserts a listing (status=pending) plus its attributes in one tx.
func (s *Service) Create(ctx context.Context, in CreateInput) (*Detail, error) {
	fields, err := s.store.ListResolvedFieldsForCategory(ctx, in.CategoryID)
	if err != nil {
		return nil, err
	}
	fieldMap := make(map[int64]sqlc.CategoryField, len(fields))
	for _, f := range fields {
		fieldMap[f.ID] = f
	}

	attrParams, err := buildAttributes(in.Attributes, fieldMap)
	if err != nil {
		return nil, err
	}
	if err := requireAttributes(fields, in.Attributes); err != nil {
		return nil, err
	}

	pt := sqlc.PriceType(in.PriceType)
	if !pt.Valid() {
		pt = sqlc.PriceTypeFixed
	}
	slug := ptr(slugify(in.Title))

	var created sqlc.Listing
	err = s.store.WithTx(ctx, func(q *sqlc.Queries) error {
		l, err := q.CreateListing(ctx, sqlc.CreateListingParams{
			UserID:         in.UserID,
			CategoryID:     in.CategoryID,
			LocationID:     in.LocationID,
			Title:          in.Title,
			Slug:           slug,
			Description:    in.Description,
			Price:          pgutil.NumericFromFloatPtr(in.Price),
			Currency:       "OMR",
			PriceType:      pt,
			ContactPhone:   in.ContactPhone,
			WhatsappNumber: in.WhatsappNumber,
			Status:         sqlc.ListingStatusPending,
		})
		if err != nil {
			return err
		}
		created = l
		for _, a := range attrParams {
			a.ListingID = l.ID
			if err := q.UpsertListingAttribute(ctx, a); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, created.ID, false)
}

// Get returns full listing detail. When incrementViews is true the view
// counter is bumped (used only for public detail views).
func (s *Service) Get(ctx context.Context, id int64, incrementViews bool) (*Detail, error) {
	row, err := s.store.GetListingByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	images, err := s.store.ListImagesByListing(ctx, id)
	if err != nil {
		return nil, err
	}
	attrs, err := s.store.ListAttributesByListing(ctx, id)
	if err != nil {
		return nil, err
	}
	if incrementViews {
		_ = s.store.IncrementListingViews(ctx, id)
	}
	return &Detail{Listing: row, Images: images, Attributes: attrs}, nil
}

// Update modifies a listing the actor owns.
func (s *Service) Update(ctx context.Context, actorID, id int64, in CreateInput) (*Detail, error) {
	if err := s.assertOwner(ctx, actorID, id); err != nil {
		return nil, err
	}
	pt := sqlc.PriceType(in.PriceType)
	if !pt.Valid() {
		pt = sqlc.PriceTypeFixed
	}

	fields, err := s.store.ListResolvedFieldsForCategory(ctx, in.CategoryID)
	if err != nil {
		return nil, err
	}
	fieldMap := make(map[int64]sqlc.CategoryField, len(fields))
	for _, f := range fields {
		fieldMap[f.ID] = f
	}
	attrParams, err := buildAttributes(in.Attributes, fieldMap)
	if err != nil {
		return nil, err
	}

	err = s.store.WithTx(ctx, func(q *sqlc.Queries) error {
		if _, err := q.UpdateListing(ctx, sqlc.UpdateListingParams{
			ID:             id,
			CategoryID:     in.CategoryID,
			LocationID:     in.LocationID,
			Title:          in.Title,
			Slug:           ptr(slugify(in.Title)),
			Description:    in.Description,
			Price:          pgutil.NumericFromFloatPtr(in.Price),
			PriceType:      pt,
			ContactPhone:   in.ContactPhone,
			WhatsappNumber: in.WhatsappNumber,
		}); err != nil {
			return err
		}
		if err := q.DeleteAttributesByListing(ctx, id); err != nil {
			return err
		}
		for _, a := range attrParams {
			a.ListingID = id
			if err := q.UpsertListingAttribute(ctx, a); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, id, false)
}

// Delete soft-deletes a listing the actor owns.
func (s *Service) Delete(ctx context.Context, actorID, id int64) error {
	if err := s.assertOwner(ctx, actorID, id); err != nil {
		return err
	}
	return s.store.SoftDeleteListing(ctx, id)
}

// Search runs the public filtered search plus a total count.
func (s *Service) Search(ctx context.Context, p sqlc.SearchListingsParams, cp sqlc.CountListingsParams) (*SearchResult, error) {
	items, err := s.store.SearchListings(ctx, p)
	if err != nil {
		return nil, err
	}
	total, err := s.store.CountListings(ctx, cp)
	if err != nil {
		return nil, err
	}
	return &SearchResult{Items: items, Total: total}, nil
}

// ListByUser returns all of a user's listings (any status).
func (s *Service) ListByUser(ctx context.Context, userID int64) ([]sqlc.ListListingsByUserRow, error) {
	return s.store.ListListingsByUser(ctx, userID)
}

// AdminList returns listings for moderation, optionally filtered by status.
func (s *Service) AdminList(ctx context.Context, status string) ([]sqlc.AdminListListingsRow, error) {
	var ns sqlc.NullListingStatus
	if status != "" {
		st := sqlc.ListingStatus(status)
		if !st.Valid() {
			return nil, domain.ErrInvalidInput
		}
		ns = sqlc.NullListingStatus{ListingStatus: st, Valid: true}
	}
	return s.store.AdminListListings(ctx, ns)
}

// SetStatus is the admin moderation action (approve/reject/etc.).
func (s *Service) SetStatus(ctx context.Context, id int64, status string) (sqlc.Listing, error) {
	st := sqlc.ListingStatus(status)
	if !st.Valid() {
		return sqlc.Listing{}, domain.ErrInvalidInput
	}
	l, err := s.store.UpdateListingStatus(ctx, sqlc.UpdateListingStatusParams{ID: id, Status: st})
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.Listing{}, domain.ErrNotFound
	}
	return l, err
}

// AddImage attaches an uploaded image to a listing the actor owns.
func (s *Service) AddImage(ctx context.Context, actorID, listingID int64, storageKey string, isPrimary bool, order int32) (sqlc.ListingImage, error) {
	if err := s.assertOwner(ctx, actorID, listingID); err != nil {
		return sqlc.ListingImage{}, err
	}
	return s.store.AddListingImage(ctx, sqlc.AddListingImageParams{
		ListingID:    listingID,
		StorageKey:   storageKey,
		IsPrimary:    isPrimary,
		DisplayOrder: order,
	})
}

// ---- favorites ----

func (s *Service) AddFavorite(ctx context.Context, userID, listingID int64) error {
	return s.store.AddFavorite(ctx, sqlc.AddFavoriteParams{UserID: userID, ListingID: listingID})
}

func (s *Service) RemoveFavorite(ctx context.Context, userID, listingID int64) error {
	return s.store.RemoveFavorite(ctx, sqlc.RemoveFavoriteParams{UserID: userID, ListingID: listingID})
}

func (s *Service) ListFavorites(ctx context.Context, userID int64) ([]sqlc.ListFavoritesByUserRow, error) {
	return s.store.ListFavoritesByUser(ctx, userID)
}

// assertOwner verifies the listing exists and belongs to actorID.
func (s *Service) assertOwner(ctx context.Context, actorID, id int64) error {
	row, err := s.store.GetListingByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	} else if err != nil {
		return err
	}
	if row.Listing.UserID != actorID {
		return domain.ErrForbidden
	}
	return nil
}

// ---- helpers ----

func buildAttributes(inputs []AttributeInput, fields map[int64]sqlc.CategoryField) ([]sqlc.UpsertListingAttributeParams, error) {
	out := make([]sqlc.UpsertListingAttributeParams, 0, len(inputs))
	for _, in := range inputs {
		f, ok := fields[in.FieldID]
		if !ok {
			return nil, domain.ErrInvalidInput
		}
		if in.Value == nil {
			continue
		}
		p := sqlc.UpsertListingAttributeParams{FieldID: in.FieldID}
		switch f.FieldType {
		case sqlc.FieldTypeNumber:
			fv, ok := toFloat(in.Value)
			if !ok {
				return nil, domain.ErrInvalidInput
			}
			p.ValueNumber = pgutil.NumericFromFloatPtr(&fv)
		case sqlc.FieldTypeBoolean:
			bv, ok := in.Value.(bool)
			if !ok {
				return nil, domain.ErrInvalidInput
			}
			p.ValueBoolean = &bv
		case sqlc.FieldTypeDate:
			sv, ok := in.Value.(string)
			if !ok {
				return nil, domain.ErrInvalidInput
			}
			var d pgtype.Date
			if err := d.Scan(sv); err != nil {
				return nil, domain.ErrInvalidInput
			}
			p.ValueDate = d
		case sqlc.FieldTypeMultiselect:
			b, err := json.Marshal(in.Value)
			if err != nil {
				return nil, domain.ErrInvalidInput
			}
			p.ValueJson = b
		default: // text, textarea, select, url
			sv, ok := in.Value.(string)
			if !ok {
				return nil, domain.ErrInvalidInput
			}
			p.ValueText = &sv
		}
		out = append(out, p)
	}
	return out, nil
}

func requireAttributes(fields []sqlc.CategoryField, inputs []AttributeInput) error {
	provided := make(map[int64]bool, len(inputs))
	for _, in := range inputs {
		if in.Value != nil {
			provided[in.FieldID] = true
		}
	}
	for _, f := range fields {
		if f.IsRequired && !provided[f.ID] {
			return domain.ErrInvalidInput
		}
	}
	return nil
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	}
	return 0, false
}

func ptr[T any](v T) *T { return &v }

// slugify produces a URL-friendly slug, keeping unicode letters/digits
// (so Arabic titles remain meaningful) and collapsing separators to '-'.
func slugify(s string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash && b.Len() > 0 {
				b.WriteRune('-')
				prevDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}
