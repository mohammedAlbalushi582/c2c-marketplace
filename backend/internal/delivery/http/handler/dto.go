package handler

import (
	"encoding/json"
	"time"

	"github.com/alamjad/marketplace/internal/platform/pgutil"
	"github.com/alamjad/marketplace/internal/repository/postgres/sqlc"
	"github.com/alamjad/marketplace/internal/usecase/listing"
)

// ---- users / auth ----

type UserDTO struct {
	ID             int64   `json:"id"`
	Email          string  `json:"email"`
	Username       *string `json:"username"`
	FullName       string  `json:"full_name"`
	Phone          *string `json:"phone"`
	WhatsappNumber *string `json:"whatsapp_number"`
	Role           string  `json:"role"`
}

func toUserDTO(u sqlc.User) UserDTO {
	return UserDTO{
		ID: u.ID, Email: u.Email, Username: u.Username, FullName: u.FullName,
		Phone: u.Phone, WhatsappNumber: u.WhatsappNumber, Role: string(u.Role),
	}
}

type AuthResponse struct {
	User         UserDTO `json:"user"`
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	ExpiresIn    int64   `json:"expires_in"`
}

// ---- categories / fields / locations ----

type CategoryDTO struct {
	ID           int64   `json:"id"`
	ParentID     *int64  `json:"parent_id"`
	Slug         string  `json:"slug"`
	NameAr       string  `json:"name_ar"`
	NameEn       *string `json:"name_en"`
	Icon         *string `json:"icon"`
	DisplayOrder int32   `json:"display_order"`
}

func toCategoryDTO(c sqlc.Category) CategoryDTO {
	return CategoryDTO{
		ID: c.ID, ParentID: c.ParentID, Slug: c.Slug, NameAr: c.NameAr,
		NameEn: c.NameEn, Icon: c.Icon, DisplayOrder: c.DisplayOrder,
	}
}

type FieldDTO struct {
	ID           int64           `json:"id"`
	FieldKey     string          `json:"field_key"`
	LabelAr      string          `json:"label_ar"`
	LabelEn      *string         `json:"label_en"`
	FieldType    string          `json:"field_type"`
	Unit         *string         `json:"unit"`
	Options      json.RawMessage `json:"options,omitempty"`
	IsRequired   bool            `json:"is_required"`
	IsFilterable bool            `json:"is_filterable"`
	DisplayOrder int32           `json:"display_order"`
}

func toFieldDTO(f sqlc.CategoryField) FieldDTO {
	return FieldDTO{
		ID: f.ID, FieldKey: f.FieldKey, LabelAr: f.LabelAr, LabelEn: f.LabelEn,
		FieldType: string(f.FieldType), Unit: f.Unit, Options: json.RawMessage(f.Options),
		IsRequired: f.IsRequired, IsFilterable: f.IsFilterable, DisplayOrder: f.DisplayOrder,
	}
}

type LocationDTO struct {
	ID       int64   `json:"id"`
	ParentID *int64  `json:"parent_id"`
	Type     string  `json:"type"`
	NameAr   string  `json:"name_ar"`
	NameEn   *string `json:"name_en"`
	Slug     string  `json:"slug"`
}

func toLocationDTO(l sqlc.Location) LocationDTO {
	return LocationDTO{
		ID: l.ID, ParentID: l.ParentID, Type: string(l.Type),
		NameAr: l.NameAr, NameEn: l.NameEn, Slug: l.Slug,
	}
}

// ---- listings ----

type ListingCardDTO struct {
	ID             int64      `json:"id"`
	Title          string     `json:"title"`
	Slug           *string    `json:"slug"`
	Price          *float64   `json:"price"`
	Currency       string     `json:"currency"`
	PriceType      string     `json:"price_type"`
	CategorySlug   *string    `json:"category_slug,omitempty"`
	CategoryNameAr *string    `json:"category_name_ar,omitempty"`
	LocationNameAr *string    `json:"location_name_ar,omitempty"`
	PrimaryImage   *string    `json:"primary_image"`
	Status         string     `json:"status"`
	IsFeatured     bool       `json:"is_featured"`
	CreatedAt      time.Time  `json:"created_at"`
}

type ImageDTO struct {
	ID        int64  `json:"id"`
	URL       string `json:"url"`
	IsPrimary bool   `json:"is_primary"`
}

type AttributeDTO struct {
	FieldID   int64  `json:"field_id"`
	FieldKey  string `json:"field_key"`
	LabelAr   string `json:"label_ar"`
	FieldType string `json:"field_type"`
	Unit      *string `json:"unit"`
	Value     any    `json:"value"`
	RawValue  any    `json:"raw_value"`
}

type ListingDetailDTO struct {
	ID             int64          `json:"id"`
	UserID         int64          `json:"user_id"`
	Title          string         `json:"title"`
	Slug           *string        `json:"slug"`
	Description    string         `json:"description"`
	Price          *float64       `json:"price"`
	Currency       string         `json:"currency"`
	PriceType      string         `json:"price_type"`
	ContactPhone   *string        `json:"contact_phone"`
	WhatsappNumber *string        `json:"whatsapp_number"`
	Status         string         `json:"status"`
	ViewsCount     int64          `json:"views_count"`
	CategoryID     int64          `json:"category_id"`
	LocationID     *int64         `json:"location_id"`
	CategorySlug   string         `json:"category_slug"`
	CategoryNameAr string         `json:"category_name_ar"`
	LocationNameAr *string        `json:"location_name_ar"`
	CreatedAt      time.Time      `json:"created_at"`
	Images         []ImageDTO     `json:"images"`
	Attributes     []AttributeDTO `json:"attributes"`
}

// urlFn resolves a storage key to a public URL.
type urlFn func(key string) string

func toDetailDTO(d *listing.Detail, url urlFn) ListingDetailDTO {
	l := d.Listing.Listing
	out := ListingDetailDTO{
		ID: l.ID, UserID: l.UserID, Title: l.Title, Slug: l.Slug, Description: l.Description,
		Price: pgutil.NumericToFloatPtr(l.Price), Currency: l.Currency,
		PriceType: string(l.PriceType), ContactPhone: l.ContactPhone,
		WhatsappNumber: l.WhatsappNumber, Status: string(l.Status), ViewsCount: l.ViewsCount,
		CategoryID: l.CategoryID, LocationID: l.LocationID, CategorySlug: d.Listing.CategorySlug,
		CategoryNameAr: d.Listing.CategoryNameAr, LocationNameAr: d.Listing.LocationNameAr,
		CreatedAt: l.CreatedAt,
		Images:    make([]ImageDTO, 0, len(d.Images)),
		Attributes: make([]AttributeDTO, 0, len(d.Attributes)),
	}
	for _, img := range d.Images {
		out.Images = append(out.Images, ImageDTO{ID: img.ID, URL: url(img.StorageKey), IsPrimary: img.IsPrimary})
	}
	for _, a := range d.Attributes {
		out.Attributes = append(out.Attributes, AttributeDTO{
			FieldID: a.FieldID, FieldKey: a.FieldKey, LabelAr: a.LabelAr, FieldType: string(a.FieldType),
			Unit: a.Unit, Value: attrValue(a), RawValue: attrRawValue(a),
		})
	}
	return out
}

type fieldOption struct {
	Value   string `json:"value"`
	LabelAr string `json:"label_ar"`
}

// optionLabels maps stored option values to their Arabic labels.
func optionLabels(raw []byte) map[string]string {
	m := map[string]string{}
	var opts []fieldOption
	if len(raw) > 0 && json.Unmarshal(raw, &opts) == nil {
		for _, o := range opts {
			m[o.Value] = o.LabelAr
		}
	}
	return m
}

// attrRawValue returns the stored value as-is. Select and multiselect values
// are option keys, which attrValue swaps for human labels — forms that write
// the value back (the edit form) need the key, not the label.
func attrRawValue(a sqlc.ListAttributesByListingRow) any {
	switch a.FieldType {
	case sqlc.FieldTypeSelect:
		return a.ValueText
	case sqlc.FieldTypeMultiselect:
		var vals []string
		if len(a.ValueJson) == 0 || json.Unmarshal(a.ValueJson, &vals) != nil {
			return nil
		}
		return vals
	default:
		return attrValue(a)
	}
}

func attrValue(a sqlc.ListAttributesByListingRow) any {
	switch a.FieldType {
	case sqlc.FieldTypeNumber:
		return pgutil.NumericToFloatPtr(a.ValueNumber)
	case sqlc.FieldTypeBoolean:
		return a.ValueBoolean
	case sqlc.FieldTypeDate:
		if a.ValueDate.Valid {
			return a.ValueDate.Time.Format("2006-01-02")
		}
		return nil
	case sqlc.FieldTypeMultiselect:
		var vals []string
		if len(a.ValueJson) == 0 || json.Unmarshal(a.ValueJson, &vals) != nil {
			return nil
		}
		labels := optionLabels(a.Options)
		out := make([]string, 0, len(vals))
		for _, v := range vals {
			if l, ok := labels[v]; ok {
				out = append(out, l)
			} else {
				out = append(out, v)
			}
		}
		return out
	case sqlc.FieldTypeSelect:
		if a.ValueText == nil {
			return nil
		}
		if l, ok := optionLabels(a.Options)[*a.ValueText]; ok {
			return l
		}
		return a.ValueText
	default:
		return a.ValueText
	}
}

func cardFromSearch(r sqlc.SearchListingsRow, url urlFn) ListingCardDTO {
	c := ListingCardDTO{
		ID: r.ID, Title: r.Title, Slug: r.Slug, Price: pgutil.NumericToFloatPtr(r.Price),
		Currency: r.Currency, PriceType: string(r.PriceType), Status: string(r.Status),
		IsFeatured: r.IsFeatured, CreatedAt: r.CreatedAt,
		CategorySlug: &r.CategorySlug, CategoryNameAr: &r.CategoryNameAr, LocationNameAr: r.LocationNameAr,
	}
	c.PrimaryImage = imageURL(r.PrimaryImage, url)
	return c
}

func cardFromUserRow(r sqlc.ListListingsByUserRow, url urlFn) ListingCardDTO {
	c := ListingCardDTO{
		ID: r.ID, Title: r.Title, Slug: r.Slug, Price: pgutil.NumericToFloatPtr(r.Price),
		Currency: r.Currency, PriceType: string(r.PriceType), Status: string(r.Status),
		IsFeatured: r.IsFeatured, CreatedAt: r.CreatedAt,
	}
	c.PrimaryImage = imageURL(r.PrimaryImage, url)
	return c
}

func cardFromFavRow(r sqlc.ListFavoritesByUserRow, url urlFn) ListingCardDTO {
	c := ListingCardDTO{
		ID: r.ID, Title: r.Title, Slug: r.Slug, Price: pgutil.NumericToFloatPtr(r.Price),
		Currency: r.Currency, PriceType: string(r.PriceType), Status: string(r.Status),
		IsFeatured: r.IsFeatured, CreatedAt: r.CreatedAt,
	}
	c.PrimaryImage = imageURL(r.PrimaryImage, url)
	return c
}

func cardFromAdminRow(r sqlc.AdminListListingsRow, url urlFn) ListingCardDTO {
	c := ListingCardDTO{
		ID: r.ID, Title: r.Title, Slug: r.Slug, Price: pgutil.NumericToFloatPtr(r.Price),
		Currency: r.Currency, PriceType: string(r.PriceType), Status: string(r.Status),
		IsFeatured: r.IsFeatured, CreatedAt: r.CreatedAt,
		CategoryNameAr: &r.CategoryNameAr,
	}
	c.PrimaryImage = imageURL(r.PrimaryImage, url)
	return c
}

// imageURL resolves a (possibly empty) storage key to an optional public URL.
func imageURL(key string, url urlFn) *string {
	if key == "" {
		return nil
	}
	u := url(key)
	return &u
}
