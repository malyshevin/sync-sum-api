package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

type CounterService interface {
	Increment(ctx context.Context) (int64, error)
	Get(ctx context.Context) (int64, error)
}

type CounterHandler struct {
	svc    CounterService
	logger *slog.Logger
	v      *validator.Validate
}

func NewCounterHandler(svc CounterService, logger *slog.Logger) *CounterHandler {
	h := &CounterHandler{svc: svc, logger: logger, v: validator.New()}
	if err := h.v.Var(svc, "required"); err != nil {
		panic("CounterHandler: svc is required")
	}
	return h
}

// RegisterRoutes wires HTTP endpoints.
func RegisterRoutes(r chi.Router, h *CounterHandler) {
	r.Post("/counter/increment", h.increment)
	r.Get("/counter", h.get)
}

func (h *CounterHandler) increment(w http.ResponseWriter, req *http.Request) {
	if h.logger != nil {
		h.logger.Info("api increment")
	}
	val, err := h.svc.Increment(req.Context())
	if err != nil {
		http.Error(w, errors.Cause(err).Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"value": val})
}

func (h *CounterHandler) get(w http.ResponseWriter, req *http.Request) {
	if h.logger != nil {
		h.logger.Debug("api get")
	}
	val, err := h.svc.Get(req.Context())
	if err != nil {
		http.Error(w, errors.Cause(err).Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"value": val})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
