package entities

import (
	"testing"

	"bookshop/backend/internal/domain/shared"

	"github.com/stretchr/testify/require"
)

func TestPublisherValidationReason(t *testing.T) {
	p := Publisher{Name: " \t"}
	err := p.Validate()
	require.ErrorIs(t, err, shared.ErrInvalid)
	require.ErrorIs(t, err, ErrNameRequired)
}

func TestPublisherValidate(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name, input, want string
		wantErr           error
	}{
		{"trimmed", " Publisher ", "Publisher", nil},
		{"invalid", "  ", "", ErrNameRequired},
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
