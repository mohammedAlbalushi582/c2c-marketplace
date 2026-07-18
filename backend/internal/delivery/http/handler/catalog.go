package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/alamjad/marketplace/internal/repository/postgres/sqlc"
	"github.com/alamjad/marketplace/internal/usecase/catalog"
)

type CatalogHandler struct {
	svc *catalog.Service
}

func NewCatalogHandler(svc *catalog.Service) *CatalogHandler {
	return &CatalogHandler{svc: svc}
}

func (h *CatalogHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.svc.ListCategories(r.Context())
	if err != nil {
		Error(w, err)
		return
	}
	out := make([]CategoryDTO, 0, len(cats))
	for _, c := range cats {
		out = append(out, toCategoryDTO(c))
	}
	JSON(w, http.StatusOK, out)
}

func (h *CatalogHandler) GetCategoryBySlug(w http.ResponseWriter, r *http.Request) {
	c, err := h.svc.GetCategoryBySlug(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, toCategoryDTO(c))
}

func (h *CatalogHandler) GetCategoryFields(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		BadRequest(w, "invalid category id")
		return
	}
	fields, err := h.svc.ResolvedFields(r.Context(), id)
	if err != nil {
		Error(w, err)
		return
	}
	out := make([]FieldDTO, 0, len(fields))
	for _, f := range fields {
		out = append(out, toFieldDTO(f))
	}
	JSON(w, http.StatusOK, out)
}

func (h *CatalogHandler) ListLocations(w http.ResponseWriter, r *http.Request) {
	locs, err := h.svc.ListAllLocations(r.Context())
	if err != nil {
		Error(w, err)
		return
	}
	out := make([]LocationDTO, 0, len(locs))
	for _, l := range locs {
		out = append(out, toLocationDTO(l))
	}
	JSON(w, http.StatusOK, out)
}

// ---- admin ----

type categoryRequest struct {
	ParentID     *int64  `json:"parent_id"`
	Slug         string  `json:"slug" validate:"required"`
	NameAr       string  `json:"name_ar" validate:"required"`
	NameEn       *string `json:"name_en"`
	Icon         *string `json:"icon"`
	DisplayOrder int32   `json:"display_order"`
	IsActive     bool    `json:"is_active"`
}

func (h *CatalogHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req categoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "invalid json")
		return
	}
	c, err := h.svc.CreateCategory(r.Context(), sqlc.CreateCategoryParams{
		ParentID: req.ParentID, Slug: req.Slug, NameAr: req.NameAr, NameEn: req.NameEn,
		Icon: req.Icon, DisplayOrder: req.DisplayOrder, IsActive: req.IsActive,
	})
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusCreated, toCategoryDTO(c))
}

func (h *CatalogHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		BadRequest(w, "invalid category id")
		return
	}
	var req categoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "invalid json")
		return
	}
	c, err := h.svc.UpdateCategory(r.Context(), sqlc.UpdateCategoryParams{
		ID: id, ParentID: req.ParentID, Slug: req.Slug, NameAr: req.NameAr, NameEn: req.NameEn,
		Icon: req.Icon, DisplayOrder: req.DisplayOrder, IsActive: req.IsActive,
	})
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, toCategoryDTO(c))
}

func (h *CatalogHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		BadRequest(w, "invalid category id")
		return
	}
	if err := h.svc.DeleteCategory(r.Context(), id); err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusNoContent, nil)
}

type fieldRequest struct {
	FieldKey     string          `json:"field_key" validate:"required"`
	LabelAr      string          `json:"label_ar" validate:"required"`
	LabelEn      *string         `json:"label_en"`
	FieldType    string          `json:"field_type" validate:"required"`
	Unit         *string         `json:"unit"`
	Options      json.RawMessage `json:"options"`
	IsRequired   bool            `json:"is_required"`
	IsFilterable bool            `json:"is_filterable"`
	DisplayOrder int32           `json:"display_order"`
}

func (h *CatalogHandler) CreateField(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		BadRequest(w, "invalid category id")
		return
	}
	var req fieldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "invalid json")
		return
	}
	ft := sqlc.FieldType(req.FieldType)
	if !ft.Valid() {
		BadRequest(w, "invalid field_type")
		return
	}
	f, err := h.svc.CreateField(r.Context(), sqlc.CreateCategoryFieldParams{
		CategoryID: id, FieldKey: req.FieldKey, LabelAr: req.LabelAr, LabelEn: req.LabelEn,
		FieldType: ft, Unit: req.Unit, Options: []byte(req.Options),
		IsRequired: req.IsRequired, IsFilterable: req.IsFilterable, DisplayOrder: req.DisplayOrder,
	})
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusCreated, toFieldDTO(f))
}
