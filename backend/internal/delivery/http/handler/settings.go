package handler

import (
	"encoding/json"
	"net/http"

	settingsuc "github.com/alamjad/marketplace/internal/usecase/settings"
)

type SettingsHandler struct {
	svc *settingsuc.Service
}

func NewSettingsHandler(svc *settingsuc.Service) *SettingsHandler {
	return &SettingsHandler{svc: svc}
}

// Get returns all app settings. GET /api/v1/admin/settings
func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	m, err := h.svc.All(r.Context())
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, m)
}

// Update upserts a set of settings. PATCH /api/v1/admin/settings
// Body: {"listing_fee_tier2":"2.000", "listing_duration_days":"30", ...}
func (h *SettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	var kv map[string]string
	if err := json.NewDecoder(r.Body).Decode(&kv); err != nil {
		BadRequest(w, "invalid json")
		return
	}
	if err := h.svc.Update(r.Context(), kv); err != nil {
		Error(w, err)
		return
	}
	m, err := h.svc.All(r.Context())
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, m)
}
