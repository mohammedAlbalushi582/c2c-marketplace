package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/alamjad/marketplace/internal/delivery/http/middleware"
	"github.com/alamjad/marketplace/internal/platform/pgutil"
	"github.com/alamjad/marketplace/internal/platform/storage"
	"github.com/alamjad/marketplace/internal/repository/postgres/sqlc"
	listinguc "github.com/alamjad/marketplace/internal/usecase/listing"
)

type ListingHandler struct {
	svc   *listinguc.Service
	store storage.Storage
}

func NewListingHandler(svc *listinguc.Service, store storage.Storage) *ListingHandler {
	return &ListingHandler{svc: svc, store: store}
}

func (h *ListingHandler) url(key string) string { return h.store.PublicURL(key) }

// actorOf adapts the authenticated principal to the usecase-layer actor.
func actorOf(u middleware.AuthUser) listinguc.Actor {
	return listinguc.Actor{ID: u.ID, Role: u.Role}
}

// ---- search ----

type searchResponse struct {
	Items    []ListingCardDTO `json:"items"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

func (h *ListingHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page := atoiDefault(q.Get("page"), 1)
	if page < 1 {
		page = 1
	}
	size := atoiDefault(q.Get("page_size"), 20)
	if size < 1 || size > 50 {
		size = 20
	}

	categoryID := parseInt64Ptr(q.Get("category_id"))
	locationID := parseInt64Ptr(q.Get("location_id"))
	keyword := strPtr(q.Get("keyword"))
	minPrice := pgutil.NumericFromFloatPtr(parseFloatPtr(q.Get("min_price")))
	maxPrice := pgutil.NumericFromFloatPtr(parseFloatPtr(q.Get("max_price")))

	res, err := h.svc.Search(r.Context(),
		sqlc.SearchListingsParams{
			CategoryID: categoryID, LocationID: locationID,
			MinPrice: minPrice, MaxPrice: maxPrice, Keyword: keyword,
			Lim: int32(size), Off: int32((page - 1) * size),
		},
		sqlc.CountListingsParams{
			CategoryID: categoryID, LocationID: locationID,
			MinPrice: minPrice, MaxPrice: maxPrice, Keyword: keyword,
		},
	)
	if err != nil {
		Error(w, err)
		return
	}

	items := make([]ListingCardDTO, 0, len(res.Items))
	for _, it := range res.Items {
		items = append(items, cardFromSearch(it, h.url))
	}
	JSON(w, http.StatusOK, searchResponse{Items: items, Total: res.Total, Page: page, PageSize: size})
}

func (h *ListingHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		BadRequest(w, "invalid listing id")
		return
	}
	d, err := h.svc.Get(r.Context(), id, true)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, toDetailDTO(d, h.url))
}

// ---- create / update / delete ----

type attributeReq struct {
	FieldID int64 `json:"field_id"`
	Value   any   `json:"value"`
}

type listingRequest struct {
	CategoryID     int64          `json:"category_id" validate:"required"`
	LocationID     *int64         `json:"location_id"`
	Title          string         `json:"title" validate:"required"`
	Description    string         `json:"description" validate:"required"`
	Price          *float64       `json:"price"`
	PriceType      string         `json:"price_type"`
	ContactPhone   *string        `json:"contact_phone"`
	WhatsappNumber *string        `json:"whatsapp_number"`
	Attributes     []attributeReq `json:"attributes"`
}

func (req listingRequest) toInput(userID int64) listinguc.CreateInput {
	attrs := make([]listinguc.AttributeInput, 0, len(req.Attributes))
	for _, a := range req.Attributes {
		attrs = append(attrs, listinguc.AttributeInput{FieldID: a.FieldID, Value: a.Value})
	}
	return listinguc.CreateInput{
		UserID: userID, CategoryID: req.CategoryID, LocationID: req.LocationID,
		Title: req.Title, Description: req.Description, Price: req.Price,
		PriceType: req.PriceType, ContactPhone: req.ContactPhone,
		WhatsappNumber: req.WhatsappNumber, Attributes: attrs,
	}
}

func (h *ListingHandler) Create(w http.ResponseWriter, r *http.Request) {
	u, _ := middleware.UserFromContext(r.Context())
	var req listingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "invalid json")
		return
	}
	if req.CategoryID == 0 || req.Title == "" || req.Description == "" {
		BadRequest(w, "category_id, title and description are required")
		return
	}
	d, err := h.svc.Create(r.Context(), req.toInput(u.ID))
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusCreated, toDetailDTO(d, h.url))
}

func (h *ListingHandler) Update(w http.ResponseWriter, r *http.Request) {
	u, _ := middleware.UserFromContext(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		BadRequest(w, "invalid listing id")
		return
	}
	var req listingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "invalid json")
		return
	}
	d, err := h.svc.Update(r.Context(), actorOf(u), id, req.toInput(u.ID))
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, toDetailDTO(d, h.url))
}

func (h *ListingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	u, _ := middleware.UserFromContext(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		BadRequest(w, "invalid listing id")
		return
	}
	if err := h.svc.Delete(r.Context(), actorOf(u), id); err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusNoContent, nil)
}

// ---- images ----

func (h *ListingHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	u, _ := middleware.UserFromContext(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		BadRequest(w, "invalid listing id")
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB
		BadRequest(w, "invalid multipart form")
		return
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		BadRequest(w, "missing image file")
		return
	}
	defer file.Close()

	isPrimary := r.FormValue("is_primary") == "true"
	key := fmt.Sprintf("listings/%d/%d_%s", id, time.Now().UnixNano(), header.Filename)
	if _, err := h.store.Save(r.Context(), key, file); err != nil {
		Error(w, err)
		return
	}
	img, err := h.svc.AddImage(r.Context(), actorOf(u), id, key, isPrimary, 0)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusCreated, ImageDTO{ID: img.ID, URL: h.url(img.StorageKey), IsPrimary: img.IsPrimary})
}

// ---- current user ----

func (h *ListingHandler) MyListings(w http.ResponseWriter, r *http.Request) {
	u, _ := middleware.UserFromContext(r.Context())
	rows, err := h.svc.ListByUser(r.Context(), u.ID)
	if err != nil {
		Error(w, err)
		return
	}
	out := make([]ListingCardDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, cardFromUserRow(row, h.url))
	}
	JSON(w, http.StatusOK, out)
}

func (h *ListingHandler) MyFavorites(w http.ResponseWriter, r *http.Request) {
	u, _ := middleware.UserFromContext(r.Context())
	rows, err := h.svc.ListFavorites(r.Context(), u.ID)
	if err != nil {
		Error(w, err)
		return
	}
	out := make([]ListingCardDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, cardFromFavRow(row, h.url))
	}
	JSON(w, http.StatusOK, out)
}

func (h *ListingHandler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	u, _ := middleware.UserFromContext(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		BadRequest(w, "invalid listing id")
		return
	}
	if err := h.svc.AddFavorite(r.Context(), u.ID, id); err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusNoContent, nil)
}

func (h *ListingHandler) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	u, _ := middleware.UserFromContext(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		BadRequest(w, "invalid listing id")
		return
	}
	if err := h.svc.RemoveFavorite(r.Context(), u.ID, id); err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusNoContent, nil)
}

// ---- admin ----

func (h *ListingHandler) AdminList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page := atoiDefault(q.Get("page"), 1)
	if page < 1 {
		page = 1
	}
	size := atoiDefault(q.Get("page_size"), 20)
	if size < 1 || size > 100 {
		size = 20
	}

	res, err := h.svc.AdminList(r.Context(), q.Get("status"), q.Get("keyword"), int32(size), int32((page-1)*size))
	if err != nil {
		Error(w, err)
		return
	}
	items := make([]ListingCardDTO, 0, len(res.Items))
	for _, row := range res.Items {
		items = append(items, cardFromAdminRow(row, h.url))
	}
	JSON(w, http.StatusOK, searchResponse{Items: items, Total: res.Total, Page: page, PageSize: size})
}

type statusRequest struct {
	Status string `json:"status" validate:"required"`
}

func (h *ListingHandler) SetStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		BadRequest(w, "invalid listing id")
		return
	}
	var req statusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "invalid json")
		return
	}
	l, err := h.svc.SetStatus(r.Context(), id, req.Status)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"id": l.ID, "status": string(l.Status)})
}

// ---- query helpers ----

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func parseInt64Ptr(s string) *int64 {
	if s == "" {
		return nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil
	}
	return &n
}

func parseFloatPtr(s string) *float64 {
	if s == "" {
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &f
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
