package logcontext

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestResolveID(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		value string
		keep  bool
	}{
		{name: "canonical v4", value: "4e643615-2a28-40a5-8e7d-fbe6dd2d498d", keep: true},
		{name: "missing"},
		{name: "malformed", value: "not-an-id"},
		{name: "non-RFC4122 variant", value: "01234567-89ab-4cde-0123-456789abcdef"},
		{name: "v1", value: "4e643615-2a28-10a5-8e7d-fbe6dd2d498d"},
		{name: "noncanonical", value: "4e643615-2a28-40a5-8e7d-fbe6dd2d498d "},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual := ResolveID(tc.value)
			id, err := uuid.Parse(actual)
			require.NoError(t, err)
			require.Equal(t, uuid.Version(4), id.Version())
			require.Equal(t, uuid.RFC4122, id.Variant())
			require.Equal(t, tc.keep, actual == tc.value)
		})
	}
}

func TestNewHandler(t *testing.T) {
	var output bytes.Buffer

	logger := slog.New(NewHandler(slog.NewJSONHandler(&output, nil))).
		With("service", "bookshop-api").WithGroup("detail").With("kind", "test")
	id := uuid.NewString()
	logger.InfoContext(StoreRequest(context.Background(), id, "GET", "GET /books/{id}"), "completed")

	var record struct {
		Service string `json:"service"`
		Detail  struct {
			Kind      string `json:"kind"`
			RequestID string `json:"request_id"`
		} `json:"detail"`
	}
	require.NoError(t, json.Unmarshal(output.Bytes(), &record))
	require.Equal(t, "bookshop-api", record.Service)
	require.Equal(t, "test", record.Detail.Kind)
	require.Equal(t, id, record.Detail.RequestID)
}
