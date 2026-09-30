package entities

import (
	"testing"

	"bookshop/backend/internal/domain/shared"

	"github.com/stretchr/testify/require"
)

func TestPublisherValidate(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name, input, want string
		wantErr           error
	}{
		{"trimmed", " Publisher ", "Publisher", nil},
		{"invalid", "  ", "", shared.ErrInvalid},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := Publisher{Name: tt.input}
			err := p.Validate()
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, p.Name)
		})
	}
}
