package handler

import (
	"context"
	"net/http"
	"time"

	"bookshop/backend/internal/api/rest/ogen"
)

const _readinessTimeout = 500 * time.Millisecond

func (Handler) CheckLiveness(context.Context) error {
	return nil
}

func (h Handler) CheckReadiness(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, _readinessTimeout)
	defer cancel()

	if err := h.CheckDatabase(ctx); err != nil {
		return &ogen.ErrorStatusCodeWithHeaders{
			StatusCode: http.StatusServiceUnavailable,
			Response:   ogen.Problem{Error: ogen.ProblemErrorInternalError},
		}
	}
	return nil
}
