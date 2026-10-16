package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewStartsWithInjectedLogger(t *testing.T) {
	t.Setenv("API_URL", "http://127.0.0.1:1")
	t.Setenv("WEB_ADDR", "127.0.0.1:0")
	application := New()
	require.NoError(t, application.Err())
	require.NoError(t, application.Start(context.Background()))
	require.NoError(t, application.Stop(context.Background()))
}
