package publishers

import (
	"context"
	"errors"
	"testing"

	"bookshop/backend/internal/adapters/publishers/mocks"
	"bookshop/backend/internal/domain/publishers/entities"
	"bookshop/backend/internal/domain/shared"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreatePublisher(t *testing.T) {
	t.Parallel()

	valid := entities.Publisher{Name: " Publisher "}
	normalized := entities.Publisher{Name: "Publisher"}
	unavailable := errors.New("unavailable")

	tests := []struct {
		name             string
		input, want      entities.Publisher
		repoErr, wantErr error
	}{
		{name: "normalizes input", input: valid, want: normalized},
		{name: "rejects blank name", input: entities.Publisher{Name: " "}, want: entities.Publisher{}, wantErr: shared.ErrInvalid},
		{name: "returns conflict", input: valid, repoErr: shared.ErrConflict, wantErr: shared.ErrConflict},
		{name: "returns repository error", input: valid, repoErr: unavailable, wantErr: unavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := mocks.NewMockPublishers(t)
			if tt.wantErr != shared.ErrInvalid {
				repo.EXPECT().CreatePublisher(mock.Anything, normalized).Return(tt.want, tt.repoErr).Once()
			}

			s := Service{Repository: repo}

			got, err := s.CreatePublisher(context.Background(), tt.input)

			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}

			require.Equal(t, tt.want, got)

			if tt.wantErr == shared.ErrInvalid {
				repo.AssertNotCalled(t, "CreatePublisher", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestUpdatePublisher(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	valid := entities.Publisher{ID: id, Name: " Publisher "}
	normalized := entities.Publisher{ID: id, Name: "Publisher"}
	unavailable := errors.New("unavailable")

	tests := []struct {
		name             string
		input, want      entities.Publisher
		repoErr, wantErr error
	}{
		{name: "normalizes input", input: valid, want: normalized},
		{name: "rejects blank name", input: entities.Publisher{ID: id, Name: " "}, want: entities.Publisher{ID: id}, wantErr: shared.ErrInvalid},
		{name: "returns missing", input: valid, repoErr: shared.ErrNotFound, wantErr: shared.ErrNotFound},
		{name: "returns conflict", input: valid, repoErr: shared.ErrConflict, wantErr: shared.ErrConflict},
		{name: "returns repository error", input: valid, repoErr: unavailable, wantErr: unavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := mocks.NewMockPublishers(t)
			if tt.wantErr != shared.ErrInvalid {
				repo.EXPECT().UpdatePublisher(mock.Anything, normalized).Return(tt.want, tt.repoErr).Once()
			}

			s := Service{Repository: repo}

			got, err := s.UpdatePublisher(context.Background(), tt.input)

			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}

			require.Equal(t, tt.want, got)

			if tt.wantErr == shared.ErrInvalid {
				repo.AssertNotCalled(t, "UpdatePublisher", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestListPublishers(t *testing.T) {
	t.Parallel()

	publishers := []entities.Publisher{{ID: uuid.New(), Name: "Publisher"}}
	unavailable := errors.New("unavailable")

	tests := []struct {
		name    string
		want    []entities.Publisher
		wantErr error
	}{
		{name: "returns publishers", want: publishers},
		{name: "returns empty list", want: []entities.Publisher{}},
		{name: "returns repository error", wantErr: unavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := mocks.NewMockPublishers(t)
			repo.EXPECT().ListPublishers(mock.Anything).Return(tt.want, tt.wantErr).Once()
			s := Service{Repository: repo}

			got, err := s.ListPublishers(context.Background())

			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestGetPublisher(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	publisher := entities.Publisher{ID: id, Name: "Publisher"}
	unavailable := errors.New("unavailable")

	tests := []struct {
		name    string
		want    entities.Publisher
		wantErr error
	}{
		{name: "returns publisher", want: publisher},
		{name: "returns missing", wantErr: shared.ErrNotFound},
		{name: "returns repository error", wantErr: unavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := mocks.NewMockPublishers(t)
			repo.EXPECT().GetPublisher(mock.Anything, id).Return(tt.want, tt.wantErr).Once()
			s := Service{Repository: repo}

			got, err := s.GetPublisher(context.Background(), id)

			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestDeletePublisher(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	unavailable := errors.New("unavailable")

	tests := []struct {
		name    string
		wantErr error
	}{
		{name: "deletes publisher"},
		{name: "returns missing", wantErr: shared.ErrNotFound},
		{name: "returns conflict", wantErr: shared.ErrConflict},
		{name: "returns repository error", wantErr: unavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := mocks.NewMockPublishers(t)
			repo.EXPECT().DeletePublisher(mock.Anything, id).Return(tt.wantErr).Once()
			s := Service{Repository: repo}

			err := s.DeletePublisher(context.Background(), id)

			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)

				if tt.wantErr == shared.ErrConflict {
					require.ErrorIs(t, err, entities.ErrReferenced)
				} else {
					require.NotErrorIs(t, err, entities.ErrReferenced)
				}
			}
		})
	}
}
