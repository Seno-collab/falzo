package handler

import (
	"be/internal/api/http/response"
	"context"
	"net/http"
	"time"
)

const healthCheckTimeout = 2 * time.Second

type HealthCheck func(context.Context) error

type HealthHandler struct {
	checks []HealthCheck
}

func NewHealthHandler(checks ...HealthCheck) *HealthHandler {
	return &HealthHandler{checks: checks}
}

func (h *HealthHandler) Live(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, struct {
		Status string `json:"status"`
	}{Status: "up"})
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), healthCheckTimeout)
	defer cancel()

	for _, check := range h.checks {
		if err := check(ctx); err != nil {
			response.JSON(w, http.StatusServiceUnavailable, struct {
				Status string `json:"status"`
			}{Status: "not_ready"})
			return
		}
	}
	response.OK(w, struct {
		Status string `json:"status"`
	}{Status: "ready"})
}
