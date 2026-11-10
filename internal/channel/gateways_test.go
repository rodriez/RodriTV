package channel

import (
	"context"
	"sort"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeRepository es una implementación en memoria de Repository para
// verificar que la interfaz es implementable por un fake.
type fakeRepository struct {
	channels map[uuid.UUID]Channel
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{channels: make(map[uuid.UUID]Channel)}
}

func (f *fakeRepository) List(_ context.Context) ([]Channel, error) {
	channels := make([]Channel, 0, len(f.channels))
	for _, ch := range f.channels {
		channels = append(channels, ch)
	}
	sort.Slice(channels, func(i, j int) bool { return channels[i].Number < channels[j].Number })

	return channels, nil
}

func (f *fakeRepository) GetByID(_ context.Context, id uuid.UUID) (Channel, error) {
	ch, ok := f.channels[id]
	if !ok {
		return Channel{}, ErrNotFound
	}

	return ch, nil
}

func (f *fakeRepository) Create(_ context.Context, ch Channel) (Channel, error) {
	exists, _ := f.ExistsNumber(context.Background(), ch.Number, uuid.Nil)
	if exists {
		return Channel{}, ErrConflict
	}
	f.channels[ch.ID] = ch

	return ch, nil
}

func (f *fakeRepository) Update(_ context.Context, ch Channel) (Channel, error) {
	if _, ok := f.channels[ch.ID]; !ok {
		return Channel{}, ErrNotFound
	}
	exists, _ := f.ExistsNumber(context.Background(), ch.Number, ch.ID)
	if exists {
		return Channel{}, ErrConflict
	}
	f.channels[ch.ID] = ch

	return ch, nil
}

func (f *fakeRepository) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := f.channels[id]; !ok {
		return ErrNotFound
	}
	delete(f.channels, id)

	return nil
}

func (f *fakeRepository) ExistsNumber(_ context.Context, number int, excludeID uuid.UUID) (bool, error) {
	for _, ch := range f.channels {
		if ch.Number == number && ch.ID != excludeID {
			return true, nil
		}
	}

	return false, nil
}

func TestRepository_FakeImplementsInterface(t *testing.T) {
	var repo Repository = newFakeRepository()

	ch := validChannel()
	created, err := repo.Create(context.Background(), ch)
	require.NoError(t, err)
	assert.Equal(t, ch.ID, created.ID)
}

func TestRepository_FakeGetByIDNotFound(t *testing.T) {
	var repo Repository = newFakeRepository()

	_, err := repo.GetByID(context.Background(), uuid.New())
	require.ErrorIs(t, err, ErrNotFound)
}

func TestRepository_FakeCreateConflict(t *testing.T) {
	repo := newFakeRepository()
	_, err := repo.Create(context.Background(), validChannel())
	require.NoError(t, err)

	other := validChannel()
	other.ID = uuid.New()
	_, err = repo.Create(context.Background(), other)
	require.ErrorIs(t, err, ErrConflict)
}

func TestRepository_FakeListOrderedByNumber(t *testing.T) {
	repo := newFakeRepository()
	ctx := context.Background()

	high := validChannel()
	high.Number = 30
	low := validChannel()
	low.ID = uuid.New()
	low.Number = 10

	_, err := repo.Create(ctx, high)
	require.NoError(t, err)
	_, err = repo.Create(ctx, low)
	require.NoError(t, err)

	channels, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, channels, 2)
	assert.Equal(t, 10, channels[0].Number)
	assert.Equal(t, 30, channels[1].Number)
}

func TestRepository_FakeUpdateNotFoundAndConflict(t *testing.T) {
	repo := newFakeRepository()
	ctx := context.Background()

	missing := validChannel()
	_, err := repo.Update(ctx, missing)
	require.ErrorIs(t, err, ErrNotFound)

	existing, err := repo.Create(ctx, validChannel())
	require.NoError(t, err)

	other := validChannel()
	other.ID = uuid.New()
	other.Number = 99
	other, err = repo.Create(ctx, other)
	require.NoError(t, err)

	existing.Number = 99
	_, err = repo.Update(ctx, existing)
	require.ErrorIs(t, err, ErrConflict)

	other.Name = "updated"
	updated, err := repo.Update(ctx, other)
	require.NoError(t, err)
	assert.Equal(t, "updated", updated.Name)
}

func TestRepository_FakeDeleteNotFound(t *testing.T) {
	repo := newFakeRepository()
	ctx := context.Background()

	require.ErrorIs(t, repo.Delete(ctx, uuid.New()), ErrNotFound)

	created, err := repo.Create(ctx, validChannel())
	require.NoError(t, err)
	require.NoError(t, repo.Delete(ctx, created.ID))
}

func TestRepository_FakeExistsNumberExcludesID(t *testing.T) {
	repo := newFakeRepository()
	ctx := context.Background()

	created, err := repo.Create(ctx, validChannel())
	require.NoError(t, err)

	exists, err := repo.ExistsNumber(ctx, created.Number, uuid.New())
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = repo.ExistsNumber(ctx, created.Number, created.ID)
	require.NoError(t, err)
	assert.False(t, exists)

	exists, err = repo.ExistsNumber(ctx, created.Number+1000, uuid.Nil)
	require.NoError(t, err)
	assert.False(t, exists)
}
