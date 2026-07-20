package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/alamjad/marketplace/internal/delivery/http/middleware"
	useruc "github.com/alamjad/marketplace/internal/usecase/user"
)

type UserHandler struct {
	svc *useruc.Service
}

func NewUserHandler(svc *useruc.Service) *UserHandler {
	return &UserHandler{svc: svc}
}

// AdminUserDTO is the user row shown in the admin panel. Unlike UserDTO it
// exposes status, join date, and listing count, but never the password hash.
type AdminUserDTO struct {
	ID             int64     `json:"id"`
	Email          string    `json:"email"`
	Username       *string   `json:"username"`
	FullName       string    `json:"full_name"`
	Phone          *string   `json:"phone"`
	WhatsappNumber *string   `json:"whatsapp_number"`
	Role           string    `json:"role"`
	Status         string    `json:"status"`
	ListingsCount  int64     `json:"listings_count"`
	CreatedAt      time.Time `json:"created_at"`
}

type adminUsersResponse struct {
	Items    []AdminUserDTO `json:"items"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page := atoiDefault(q.Get("page"), 1)
	if page < 1 {
		page = 1
	}
	size := atoiDefault(q.Get("page_size"), 20)
	if size < 1 || size > 100 {
		size = 20
	}

	res, err := h.svc.List(r.Context(), q.Get("keyword"), q.Get("role"), int32(size), int32((page-1)*size))
	if err != nil {
		Error(w, err)
		return
	}

	items := make([]AdminUserDTO, 0, len(res.Items))
	for _, u := range res.Items {
		items = append(items, AdminUserDTO{
			ID: u.ID, Email: u.Email, Username: u.Username, FullName: u.FullName,
			Phone: u.Phone, WhatsappNumber: u.WhatsappNumber,
			Role: string(u.Role), Status: string(u.Status),
			ListingsCount: u.ListingsCount, CreatedAt: u.CreatedAt,
		})
	}
	JSON(w, http.StatusOK, adminUsersResponse{Items: items, Total: res.Total, Page: page, PageSize: size})
}

type roleRequest struct {
	Role string `json:"role" validate:"required"`
}

func (h *UserHandler) SetRole(w http.ResponseWriter, r *http.Request) {
	actor, _ := middleware.UserFromContext(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		BadRequest(w, "invalid user id")
		return
	}
	var req roleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "invalid json")
		return
	}
	u, err := h.svc.SetRole(r.Context(), actor.ID, id, req.Role)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"id": u.ID, "role": string(u.Role)})
}

type userStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

func (h *UserHandler) SetStatus(w http.ResponseWriter, r *http.Request) {
	actor, _ := middleware.UserFromContext(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		BadRequest(w, "invalid user id")
		return
	}
	var req userStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "invalid json")
		return
	}
	u, err := h.svc.SetStatus(r.Context(), actor.ID, id, req.Status)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"id": u.ID, "status": string(u.Status)})
}
