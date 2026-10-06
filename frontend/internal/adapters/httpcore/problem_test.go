package httpcore

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProblemResponsePreservesErrorCode(t *testing.T) {
	t.Parallel()
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"book_price_negative"}`))
	}))
	t.Cleanup(api.Close)
	client := New(api.URL)
	t.Cleanup(client.Stop)

	err := client.Request(context.Background(), "POST", "/books", nil, nil)
	var problem *APIError
	require.ErrorAs(t, fmt.Errorf("save: %w", err), &problem)
	require.Equal(t, 400, problem.StatusCode)
	require.Equal(t, "book_price_negative", problem.Code)
}

func TestMalformedProblemDoesNotExposeResponseBody(t *testing.T) {
	t.Parallel()
	for _, body := range []string{"private proxy failure", `{}`, `{"error":123}`, `{"error":["book_price_negative"]}`} {
		t.Run(body, func(t *testing.T) {
			t.Parallel()
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
				_, _ = w.Write([]byte(body))
			}))
			t.Cleanup(api.Close)
			client := New(api.URL)
			t.Cleanup(client.Stop)
			err := client.Request(context.Background(), "GET", "/books", nil, nil)
			var problem *APIError
			require.ErrorAs(t, err, &problem)
			require.Equal(t, 502, problem.StatusCode)
			require.Equal(t, "invalid_response", problem.Code)
			require.NotContains(t, err.Error(), body)
		})
	}
}
