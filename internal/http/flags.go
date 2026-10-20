package httpx

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/xarop-pa-toss/flaggo/internal/model"
	"github.com/xarop-pa-toss/flaggo/internal/store"
)

// FlagsHandler serves the control-plane flag endpoints.
type FlagsHandler struct {
	store  store.Store
	logger *slog.Logger
}

// NewFlagsHandler constructs a handler.
func NewFlagsHandler(s store.Store, log *slog.Logger) *FlagsHandler {
	return &FlagsHandler{store: s, logger: log}
}

// Register wires routes into mux.
func (h *FlagsHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /flags", h.Create)
	mux.HandleFunc("GET /flags", h.List)
	mux.HandleFunc("GET /flags/{name}", h.Get)
	mux.HandleFunc("PUT /flags/{name}", h.Update)
	mux.HandleFunc("DELETE /flags/{name}", h.Delete)
	mux.HandleFunc("POST /flags/{name}/kill-switch", h.KillSwitch)
	mux.HandleFunc("PUT /flags/{name}/rollout", h.Rollout)
}

func (h *FlagsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var f model.Flag
	if err := decodeJSON(r, &f); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := f.Validate(); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.store.Create(r.Context(), f); err != nil {
		if errors.Is(err, model.ErrAlreadyExists) {
			writeErr(w, http.StatusConflict, err.Error())
			return
		}
		h.logger.Error("create flag failed", "err", err)
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	out, _ := h.store.Get(r.Context(), f.Name)
	writeJSON(w, http.StatusCreated, out)
}

func (h *FlagsHandler) List(w http.ResponseWriter, r *http.Request) {
	flags, err := h.store.List(r.Context())
	if err != nil {
		h.logger.Error("list flags failed", "err", err)
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	if flags == nil {
		flags = []model.Flag{}
	}
	writeJSON(w, http.StatusOK, flags)
}

func (h *FlagsHandler) Get(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	f, err := h.store.Get(r.Context(), name)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			writeErr(w, http.StatusNotFound, err.Error())
			return
		}
		h.logger.Error("get flag failed", "err", err)
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, f)
}

func (h *FlagsHandler) Update(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	var f model.Flag
	if err := decodeJSON(r, &f); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	f.Name = name
	if err := f.Validate(); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.store.Update(r.Context(), f); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			writeErr(w, http.StatusNotFound, err.Error())
			return
		}
		h.logger.Error("update flag failed", "err", err)
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	out, _ := h.store.Get(r.Context(), name)
	writeJSON(w, http.StatusOK, out)
}

func (h *FlagsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := h.store.Delete(r.Context(), name); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			writeErr(w, http.StatusNotFound, err.Error())
			return
		}
		h.logger.Error("delete flag failed", "err", err)
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type killSwitchReq struct {
	KillSwitch *bool `json:"kill_switch"`
}

func (h *FlagsHandler) KillSwitch(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	f, err := h.store.Get(r.Context(), name)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			writeErr(w, http.StatusNotFound, err.Error())
			return
		}
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	var req killSwitchReq
	_ = decodeJSON(r, &req)
	if req.KillSwitch != nil {
		f.KillSwitch = *req.KillSwitch
	} else {
		f.KillSwitch = !f.KillSwitch
	}
	if err := h.store.Update(r.Context(), f); err != nil {
		h.logger.Error("kill switch update failed", "err", err)
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	out, _ := h.store.Get(r.Context(), name)
	writeJSON(w, http.StatusOK, out)
}

type rolloutReq struct {
	RolloutPercent int `json:"rollout_percent"`
}

func (h *FlagsHandler) Rollout(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	f, err := h.store.Get(r.Context(), name)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			writeErr(w, http.StatusNotFound, err.Error())
			return
		}
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	var req rolloutReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	f.RolloutPercent = req.RolloutPercent
	if err := f.Validate(); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.store.Update(r.Context(), f); err != nil {
		h.logger.Error("rollout update failed", "err", err)
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	out, _ := h.store.Get(r.Context(), name)
	writeJSON(w, http.StatusOK, out)
}

func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
