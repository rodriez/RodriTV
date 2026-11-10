package channel

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rodriez/RodriTV/internal/platform"
	"github.com/rodriez/RodriTV/migrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestPostgresRepository_CRUD(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)

	ctx := context.Background()
	container, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithDatabase("rodritv_test"),
		postgres.WithUsername("rodritv"),
		postgres.WithPassword("rodritv"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	require.NoError(t, platform.MigrateUp(dsn, migrations.FS))

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	repo := NewPostgresRepository(pool)
	fixedNow := time.Date(2026, time.November, 10, 12, 0, 0, 0, time.FixedZone("test", -3*60*60))
	repo.now = func() time.Time { return fixedNow }

	first := validChannel()
	first.Number = 20
	first.Name = "Canal 20"
	first.Logo = "https://example.com/logo.png"
	first.ScheduleURL = "https://example.com/schedule"
	first.Sources[1] = &Source{
		ID:        uuid.New(),
		Slot:      2,
		URL:       "https://example.com/stream.m3u8",
		Enabled:   true,
		CreatedAt: fixedNow,
		UpdatedAt: fixedNow,
	}

	created, err := repo.Create(ctx, first)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, created.ID)
	assert.Equal(t, fixedNow.UTC(), created.CreatedAt)
	assert.Equal(t, fixedNow.UTC(), created.UpdatedAt)
	assert.Equal(t, first.Sources, created.Sources)

	second := validChannel()
	second.Number = 10
	second.Name = "Canal 10"
	_, err = repo.Create(ctx, second)
	require.NoError(t, err)

	_, err = repo.Create(ctx, first)
	assert.ErrorIs(t, err, ErrConflict)

	got, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created, got)

	channels, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, channels, 2)
	assert.Equal(t, 10, channels[0].Number)
	assert.Equal(t, 20, channels[1].Number)

	created.Name = "Canal 20 actualizado"
	created.Number = 30
	updated, err := repo.Update(ctx, created)
	require.NoError(t, err)
	assert.Equal(t, created.Name, updated.Name)
	assert.Equal(t, created.Number, updated.Number)
	assert.Equal(t, created.CreatedAt, updated.CreatedAt)
	assert.Equal(t, fixedNow.UTC(), updated.UpdatedAt)

	exists, err := repo.ExistsNumber(ctx, 10, uuid.Nil)
	require.NoError(t, err)
	assert.True(t, exists)
	exists, err = repo.ExistsNumber(ctx, 10, channels[0].ID)
	require.NoError(t, err)
	assert.False(t, exists)

	created.Number = 10
	_, err = repo.Update(ctx, created)
	assert.ErrorIs(t, err, ErrConflict)

	require.NoError(t, repo.Delete(ctx, created.ID))
	_, err = repo.GetByID(ctx, created.ID)
	assert.ErrorIs(t, err, ErrNotFound)
	assert.ErrorIs(t, repo.Delete(ctx, created.ID), ErrNotFound)

	created.ID = uuid.New()
	_, err = repo.Update(ctx, created)
	assert.ErrorIs(t, err, ErrNotFound)
}
