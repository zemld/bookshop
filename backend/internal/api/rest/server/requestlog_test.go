package server

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bookshop/backend/internal/api/rest/handler"
	"bookshop/backend/internal/requestlog"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestLogBackendRequests(t *testing.T) {
	var output bytes.Buffer

	original := slog.Default()

	slog.SetDefault(slog.New(requestlog.NewHandler(slog.NewJSONHandler(&output, nil).WithAttrs([]slog.Attr{slog.String("service", "bookshop-api")}))))
	t.Cleanup(func() { slog.SetDefault(original) })

	api := New(handler.Handler{})

	provided := uuid.NewString()

	var responseIDs []string

	for _, incoming := range []string{provided, "", "not-a-uuid", "01234567-89ab-4cde-0123-456789abcdef"} {
		req := httptest.NewRequest("GET", "/books/not-an-id?secret=private", nil)
		req.Header.Set("X-Request-ID", incoming)

		res := httptest.NewRecorder()
		api.ServeHTTP(res, req)
		actual := res.Header().Get("X-Request-ID")
		id, err := uuid.Parse(actual)
		require.NoError(t, err)
		require.Equal(t, uuid.Version(4), id.Version())
		require.Equal(t, uuid.RFC4122, id.Variant())

		responseIDs = append(responseIDs, actual)

		if incoming == provided {
			require.Equal(t, provided, actual)
		} else {
			require.NotEqual(t, incoming, actual)
		}

		require.Equal(t, http.StatusBadRequest, res.Code)
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	require.Len(t, lines, 4)

	for i, line := range lines {
		var record map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &record))
		require.Equal(t, "bookshop-api", record["service"])
		require.Equal(t, "http.request_completed", record["event"])
		require.Equal(t, http.StatusBadRequest, int(record["status_code"].(float64)))
		require.Equal(t, "/books/{id}", record["route"])
		require.Equal(t, responseIDs[i], record["request_id"])
		id, err := uuid.Parse(record["request_id"].(string))
		require.NoError(t, err)
		require.Equal(t, uuid.RFC4122, id.Variant())
		require.NotContains(t, line, "private")
		require.NotContains(t, line, "not-an-id")
	}
}
