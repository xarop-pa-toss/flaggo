package httpx

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

// HealthHandler serves /healthz and /readyz.
type HealthHandler struct {
	ready atomic.Bool
}

// NewHealthHandler returns a handler whose readiness can be toggled.
func NewHealthHandler() *HealthHandler {
	h := &HealthHandler{}
	h.ready.Store(true)
	return h
}

// SetReady toggles readiness (e.g. DB connection state).
func (h *HealthHandler) SetReady(v bool) { h.ready.Store(v) }

// IsReady reports current readiness.
func (h *HealthHandler) IsReady() bool { return h.ready.Load() }

func (h *HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *HealthHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	if !h.ready.Load() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not ready"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
