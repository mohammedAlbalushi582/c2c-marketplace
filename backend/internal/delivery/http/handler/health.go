package handler

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthHandler reports liveness and database connectivity.
type HealthHandler struct {
	pool *pgxpool.Pool
}

func NewHealthHandler(pool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{pool: pool}
}

// Health returns 200 with {"status":"ok"} when the DB is reachable, else 503.
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"status": "ok", "db": "up"}
	code := http.StatusOK

	if err := h.pool.Ping(r.Context()); err != nil {
		resp["status"] = "degraded"
		resp["db"] = "down"
		code = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(resp)
}
