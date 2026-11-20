package handler

import (
	"context"
	"net/http"
	"time"
)

const _readinessTimeout = 500 * time.Millisecond

type Health struct {
	CheckReadiness func(context.Context) error
}

func (h Health) ServeLiveness(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h Health) ServeReadiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), _readinessTimeout)
	defer cancel()

	if err := h.CheckReadiness(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
}
