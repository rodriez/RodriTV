package channel

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) List(ctx context.Context) ([]Channel, error) {
	args := m.Called(ctx)
	return args.Get(0).([]Channel), args.Error(1)
}

func (m *mockRepository) GetByID(ctx context.Context, id uuid.UUID) (Channel, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(Channel), args.Error(1)
}

func (m *mockRepository) Create(ctx context.Context, ch Channel) (Channel, error) {
	args := m.Called(ctx, ch)
	return args.Get(0).(Channel), args.Error(1)
}

func (m *mockRepository) Update(ctx context.Context, ch Channel) (Channel, error) {
	args := m.Called(ctx, ch)
	return args.Get(0).(Channel), args.Error(1)
}

func (m *mockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockRepository) ExistsNumber(ctx context.Context, number int, excludeID uuid.UUID) (bool, error) {
	args := m.Called(ctx, number, excludeID)
	return args.Bool(0), args.Error(1)
}

type mockLogoStore struct {
	mock.Mock
}

func (m *mockLogoStore) Delete(ctx context.Context, name string) error {
	return m.Called(ctx, name).Error(0)
}

func TestServiceCreate_NumberConflict(t *testing.T) {
	repo := new(mockRepository)
	ch := validChannel()
	repo.On("ExistsNumber", mock.Anything, ch.Number, uuid.Nil).Return(true, nil)

	_, err := NewService(repo, time.Now).Create(context.Background(), ch)

	require.ErrorIs(t, err, ErrConflict)
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestServiceCreate_DefaultsAndValidation(t *testing.T) {
	fixedNow := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)
	t.Run("defaults", func(t *testing.T) {
		repo := new(mockRepository)
		ch := validChannel()
		ch.SegmentSeconds = 0
		ch.Enabled = false
		ch.Sources[0] = &Source{Slot: 1, URL: "https://example.com/live.m3u8"}
		repo.On("ExistsNumber", mock.Anything, ch.Number, uuid.Nil).Return(false, nil)
		repo.On("Create", mock.Anything, mock.MatchedBy(func(created Channel) bool {
			source := created.Sources[0]
			return created.SegmentSeconds == DefaultSegmentSeconds && created.Enabled &&
				source.ID != uuid.Nil && source.ID.Version() == 4 &&
				source.CreatedAt.Equal(fixedNow) && source.UpdatedAt.Equal(fixedNow)
		})).Return(ch, nil)

		_, err := NewService(repo, func() time.Time { return fixedNow }).Create(context.Background(), ch)

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	for _, seconds := range []int{1, 11} {
		t.Run("invalid segment seconds", func(t *testing.T) {
			repo := new(mockRepository)
			ch := validChannel()
			ch.SegmentSeconds = seconds

			_, err := NewService(repo, time.Now).Create(context.Background(), ch)

			assertValidationField(t, err, "segment_seconds")
			repo.AssertNotCalled(t, "ExistsNumber", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func TestServiceUpdate_SourceIdentityAndTimestamps(t *testing.T) {
	fixedNow := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)
	repo := new(mockRepository)
	ch := validChannel()
	existingID := uuid.New()
	createdAt := fixedNow.Add(-time.Hour)
	ch.Sources[0] = &Source{ID: existingID, Slot: 1, URL: "https://example.com/one.m3u8", CreatedAt: createdAt}
	ch.Sources[1] = &Source{Slot: 2, URL: "https://example.com/two.m3u8"}
	repo.On("ExistsNumber", mock.Anything, ch.Number, ch.ID).Return(false, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(updated Channel) bool {
		existing := updated.Sources[0]
		added := updated.Sources[1]
		return existing.ID == existingID && existing.CreatedAt.Equal(createdAt) && existing.UpdatedAt.Equal(fixedNow) &&
			added.ID != uuid.Nil && added.ID.Version() == 4 && added.CreatedAt.Equal(fixedNow) && added.UpdatedAt.Equal(fixedNow)
	})).Return(ch, nil)

	_, err := NewService(repo, func() time.Time { return fixedNow }).Update(context.Background(), ch)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestServiceDelete_NotFound(t *testing.T) {
	repo := new(mockRepository)
	id := uuid.New()
	repo.On("GetByID", mock.Anything, id).Return(Channel{}, ErrNotFound)

	err := NewService(repo, time.Now).Delete(context.Background(), id)

	require.ErrorIs(t, err, ErrNotFound)
	repo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestServiceDelete_RemovesLogo(t *testing.T) {
	repo := new(mockRepository)
	logos := new(mockLogoStore)
	ch := validChannel()
	ch.Logo = "channel.png"
	repo.On("GetByID", mock.Anything, ch.ID).Return(ch, nil)
	repo.On("Delete", mock.Anything, ch.ID).Return(nil)
	logos.On("Delete", mock.Anything, ch.Logo).Return(nil)
	svc := NewService(repo, time.Now)
	svc.SetLogoStore(logos)

	require.NoError(t, svc.Delete(context.Background(), ch.ID))
	repo.AssertExpectations(t)
	logos.AssertExpectations(t)
}

func TestServiceListAndGet_PropagateErrors(t *testing.T) {
	expected := errors.New("repository error")
	repo := new(mockRepository)
	repo.On("List", mock.Anything).Return([]Channel(nil), expected)
	id := uuid.New()
	repo.On("GetByID", mock.Anything, id).Return(Channel{}, expected)
	svc := NewService(repo, time.Now)

	_, err := svc.List(context.Background())
	require.ErrorIs(t, err, expected)
	_, err = svc.Get(context.Background(), id)
	require.ErrorIs(t, err, expected)
}
