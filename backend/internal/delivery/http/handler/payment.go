package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/alamjad/marketplace/internal/delivery/http/middleware"
	"github.com/alamjad/marketplace/internal/platform/pgutil"
	"github.com/alamjad/marketplace/internal/repository/postgres/sqlc"
	paymentuc "github.com/alamjad/marketplace/internal/usecase/payment"
)

type PaymentHandler struct {
	svc *paymentuc.Service
}

func NewPaymentHandler(svc *paymentuc.Service) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

// paymentDTO is the client view of a payment (shared with listing create/pay).
type paymentDTO struct {
	ID          int64   `json:"id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Status      string  `json:"status"`
	Purpose     string  `json:"purpose"`
	CheckoutURL string  `json:"checkout_url,omitempty"`
}

func toPaymentDTO(p sqlc.Payment, checkoutURL string) paymentDTO {
	amount := 0.0
	if f := pgutil.NumericToFloatPtr(p.Amount); f != nil {
		amount = *f
	}
	return paymentDTO{
		ID: p.ID, Amount: amount, Currency: p.Currency,
		Status: string(p.Status), Purpose: p.Purpose, CheckoutURL: checkoutURL,
	}
}

// Quote returns the fee the caller would owe for their next listing.
// GET /api/v1/me/listing-fee
func (h *PaymentHandler) Quote(w http.ResponseWriter, r *http.Request) {
	u, _ := middleware.UserFromContext(r.Context())
	q, err := h.svc.QuoteListingFee(r.Context(), u.ID)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, q)
}

// Verify confirms a checkout with the gateway and releases the listing if paid.
// POST /api/v1/payments/{id}/verify
func (h *PaymentHandler) Verify(w http.ResponseWriter, r *http.Request) {
	u, _ := middleware.UserFromContext(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		BadRequest(w, "invalid payment id")
		return
	}
	res, err := h.svc.Verify(r.Context(), u.ID, id)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"status": string(res.Status), "paid": res.Paid})
}

// Get returns a payment's current state (for polling).
// GET /api/v1/payments/{id}
func (h *PaymentHandler) Get(w http.ResponseWriter, r *http.Request) {
	u, _ := middleware.UserFromContext(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		BadRequest(w, "invalid payment id")
		return
	}
	p, err := h.svc.Get(r.Context(), u.ID, id)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, toPaymentDTO(p, ""))
}
