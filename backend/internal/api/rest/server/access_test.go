package server

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"bookshop/backend/internal/logcontext"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestWrapAccess(t *testing.T) {
	for _, tc := range []struct {
		name     string
		incoming string
		flush    bool
		status   int
	}{
		{name: "preserves canonical v4", incoming: uuid.NewString(), status: http.StatusConflict},
		{name: "replaces invalid variant", incoming: "01234567-89ab-4cde-0123-456789abcdef", status: http.StatusConflict},
		{name: "flush commits implicit OK", flush: true, status: http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer

			logger := slog.New(logcontext.NewHandler(slog.NewJSONHandler(&output, nil)))
			handler := wrapAccess(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				if tc.flush {
					w.(http.Flusher).Flush()
					w.WriteHeader(http.StatusInternalServerError)

					return
				}

				w.WriteHeader(tc.status)
			}), func(*http.Request) string { return "GET /books/{id}" }, logger)

			done := make(chan struct{})

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handler.ServeHTTP(w, r)
				close(done)
			}))
			defer server.Close()

			request, err := http.NewRequest(http.MethodGet, server.URL+"/books/private?secret=hidden", nil)
			require.NoError(t, err)
			request.Header.Set("X-Request-ID", tc.incoming)
			response, err := server.Client().Do(request)
			require.NoError(t, err)
			require.NoError(t, response.Body.Close())
			<-done
			require.Equal(t, tc.status, response.StatusCode)

			var record map[string]any
			require.NoError(t, json.Unmarshal(output.Bytes(), &record))

			id := response.Header.Get("X-Request-ID")
			parsed, err := uuid.Parse(id)
			require.NoError(t, err)
			require.Equal(t, uuid.RFC4122, parsed.Variant())
			require.Equal(t, uuid.Version(4), parsed.Version())

			if tc.incoming == "01234567-89ab-4cde-0123-456789abcdef" {
				require.NotEqual(t, tc.incoming, id)
			} else if tc.incoming != "" {
				require.Equal(t, tc.incoming, id)
			}

			require.Equal(t, id, record["request_id"])
			require.Equal(t, tc.status, int(record["status_code"].(float64)))
			require.Equal(t, "GET /books/{id}", record["route"])
			require.NotContains(t, output.String(), "hidden")
			require.NotContains(t, output.String(), "private")
		})
	}
}
