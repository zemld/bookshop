package handler

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"bookshop/backend/internal/api/rest/server"
	"bookshop/backend/internal/services/publishers"

	"github.com/stretchr/testify/require"
)

func TestRejectInvalidRequest(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, method, path, body, contentType string
		want                                  int
		check                                 func(*testing.T, map[string]string)
	}{
		{"malformed ID", "GET", "/books/no-uuid", "", "", 400, checkParserError},
		{"malformed JSON", "POST", "/books", "{", "application/json", 400, checkParserError},
		{"unknown field", "POST", "/publishers", `{"unknown":1}`, "application/json", 400, checkParserError},
		{"trailing JSON", "POST", "/publishers", `{"name":"A"} {}`, "application/json", 400, checkParserError},
		{"incorrect content type", "POST", "/publishers", `{"name":"A"}`, "text/plain", 400, checkParserError},
		{"blank publisher", "POST", "/publishers", `{"name":"   "}`, "application/json", 400, func(t *testing.T, problem map[string]string) {
			require.Equal(t, map[string]string{"error": "invalid input"}, problem)
		}},
		{"absent book years", "POST", "/books", `{"author":"A","name":"B"}`, "application/json", 400, checkParserError},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			res := httptest.NewRecorder()

			server.New(Handler{Publishers: publishers.Service{}}).ServeHTTP(res, req)

			require.Equal(t, tt.want, res.Code)

			var problem map[string]string
			require.NoError(t, json.Unmarshal(res.Body.Bytes(), &problem))
			tt.check(t, problem)
		})
	}
}

func checkParserError(t *testing.T, problem map[string]string) {
	t.Helper()
	require.Equal(t, []string{"error"}, collectKeys(problem))
	require.NotEmpty(t, problem["error"])
}

func collectKeys(problem map[string]string) []string {
	result := make([]string, 0, len(problem))
	for key := range problem {
		result = append(result, key)
	}

	return result
}
