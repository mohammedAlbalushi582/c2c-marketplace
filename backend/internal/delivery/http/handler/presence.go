package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/alamjad/marketplace/internal/delivery/http/middleware"
	"github.com/alamjad/marketplace/internal/platform/presence"
)

type PresenceHandler struct {
	tracker *presence.Tracker
}

func NewPresenceHandler(tracker *presence.Tracker) *PresenceHandler {
	return &PresenceHandler{tracker: tracker}
}

type pingRequest struct {
	SessionID string `json:"session_id"`
	Path      string `json:"path"`
	Name      string `json:"name"`
}

// Ping records a heartbeat. Runs under Optional auth: a valid token attaches the
// logged-in identity; otherwise the visitor is a guest. POST /api/v1/presence/ping
func (h *PresenceHandler) Ping(w http.ResponseWriter, r *http.Request) {
	var req pingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SessionID == "" {
		BadRequest(w, "session_id required")
		return
	}

	v := presence.Visitor{
		SessionID: req.SessionID,
		Path:      req.Path,
		IP:        r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Name:      "زائر",
	}
	if u, ok := middleware.UserFromContext(r.Context()); ok {
		v.IsUser = true
		v.UserID = &u.ID
		if req.Name != "" {
			v.Name = req.Name
		} else {
			v.Name = "مستخدم"
		}
	}
	h.tracker.Touch(v)
	JSON(w, http.StatusNoContent, nil)
}

type presenceVisitorDTO struct {
	Name       string `json:"name"`
	IsUser     bool   `json:"is_user"`
	Path       string `json:"path"`
	SecondsAgo int    `json:"seconds_ago"`
}

type presenceResponse struct {
	Count    int                  `json:"count"`
	Users    int                  `json:"users"`
	Guests   int                  `json:"guests"`
	Visitors []presenceVisitorDTO `json:"visitors"`
}

// List returns the currently-online visitors for the admin panel.
// GET /api/v1/admin/presence
func (h *PresenceHandler) List(w http.ResponseWriter, r *http.Request) {
	active := h.tracker.Active()
	now := time.Now()
	resp := presenceResponse{Count: len(active), Visitors: make([]presenceVisitorDTO, 0, len(active))}
	for _, v := range active {
		if v.IsUser {
			resp.Users++
		} else {
			resp.Guests++
		}
		resp.Visitors = append(resp.Visitors, presenceVisitorDTO{
			Name:       v.Name,
			IsUser:     v.IsUser,
			Path:       v.Path,
			SecondsAgo: int(now.Sub(v.LastSeen).Seconds()),
		})
	}
	JSON(w, http.StatusOK, resp)
}
