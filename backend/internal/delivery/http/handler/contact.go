package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	contactuc "github.com/alamjad/marketplace/internal/usecase/contact"
)

type ContactHandler struct {
	svc *contactuc.Service
}

func NewContactHandler(svc *contactuc.Service) *ContactHandler {
	return &ContactHandler{svc: svc}
}

type contactMessageDTO struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Phone     *string   `json:"phone"`
	Email     *string   `json:"email"`
	Message   string    `json:"message"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

type contactRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

// Create is the public راسلنا submission. POST /api/v1/contact
func (h *ContactHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req contactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "invalid json")
		return
	}
	if req.Name == "" || req.Message == "" {
		BadRequest(w, "name and message are required")
		return
	}
	m, err := h.svc.Create(r.Context(), contactuc.CreateInput{
		Name: req.Name, Phone: req.Phone, Email: req.Email, Message: req.Message,
	})
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusCreated, toContactDTO(m.ID, req))
}

type contactListResponse struct {
	Items    []contactMessageDTO `json:"items"`
	Total    int64               `json:"total"`
	Unread   int64               `json:"unread"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

// List is the admin inbox. GET /api/v1/admin/contact-messages
func (h *ContactHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page := atoiDefault(q.Get("page"), 1)
	if page < 1 {
		page = 1
	}
	size := atoiDefault(q.Get("page_size"), 20)
	if size < 1 || size > 100 {
		size = 20
	}
	res, err := h.svc.List(r.Context(), int32(size), int32((page-1)*size))
	if err != nil {
		Error(w, err)
		return
	}
	items := make([]contactMessageDTO, 0, len(res.Items))
	for _, m := range res.Items {
		items = append(items, contactMessageDTO{
			ID: m.ID, Name: m.Name, Phone: m.Phone, Email: m.Email,
			Message: m.Message, IsRead: m.IsRead, CreatedAt: m.CreatedAt,
		})
	}
	JSON(w, http.StatusOK, contactListResponse{
		Items: items, Total: res.Total, Unread: res.Unread, Page: page, PageSize: size,
	})
}

// MarkRead marks one message read. PATCH /api/v1/admin/contact-messages/{id}/read
func (h *ContactHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		BadRequest(w, "invalid id")
		return
	}
	if err := h.svc.MarkRead(r.Context(), id); err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusNoContent, nil)
}

// Delete removes a message. DELETE /api/v1/admin/contact-messages/{id}
func (h *ContactHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		BadRequest(w, "invalid id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusNoContent, nil)
}

func toContactDTO(id int64, req contactRequest) contactMessageDTO {
	dto := contactMessageDTO{ID: id, Name: req.Name, Message: req.Message, CreatedAt: time.Now()}
	if req.Phone != "" {
		dto.Phone = &req.Phone
	}
	if req.Email != "" {
		dto.Email = &req.Email
	}
	return dto
}
