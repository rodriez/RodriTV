package platform

import (
	"context"
	"database/sql"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rodriez/RodriTV/migrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startPostgres(t *testing.T) string {
	t.Helper()
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

	return dsn
}

func openDB(t *testing.T, dsn string) *sql.DB {
	t.Helper()

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	return db
}

func queryStrings(t *testing.T, db *sql.DB, query string, args ...any) []string {
	t.Helper()

	rows, err := db.Query(query, args...)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()

	var out []string
	for rows.Next() {
		var s string
		require.NoError(t, rows.Scan(&s))
		out = append(out, s)
	}
	require.NoError(t, rows.Err())

	return out
}

func TestMigrateUp_CreatesSchema(t *testing.T) {
	dsn := startPostgres(t)

	require.NoError(t, MigrateUp(dsn, migrations.FS))

	db := openDB(t, dsn)

	tables := queryStrings(t, db,
		`SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE' ORDER BY table_name`)
	assert.Equal(t, []string{"channels", "config", "process_logs", "uptime_samples"}, tables)

	indexes := queryStrings(t, db,
		`SELECT indexname FROM pg_indexes WHERE schemaname = 'public' AND indexname LIKE 'idx_%' ORDER BY indexname`)
	assert.Equal(t, []string{"idx_process_logs_channel_ts", "idx_uptime_channel_ts"}, indexes)

	seeds := queryStrings(t, db, `SELECT key FROM config ORDER BY key`)
	assert.Equal(t, []string{
		"buffer_minutes",
		"current_channel_ttl_minutes",
		"neighbors",
		"startup_channels",
		"uptime_threshold",
	}, seeds)
}

func TestMigrateUp_NoChangeIsNil(t *testing.T) {
	dsn := startPostgres(t)

	require.NoError(t, MigrateUp(dsn, migrations.FS))
	assert.NoError(t, MigrateUp(dsn, migrations.FS))
}

func TestMigrateUp_UpDownReversible(t *testing.T) {
	dsn := startPostgres(t)

	require.NoError(t, MigrateUp(dsn, migrations.FS))

	source, err := iofs.New(migrations.FS, ".")
	require.NoError(t, err)
	m, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	require.NoError(t, err)
	defer func() { _, _ = m.Close() }()

	require.NoError(t, m.Down())

	db := openDB(t, dsn)
	tables := queryStrings(t, db,
		`SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE'`)
	assert.Empty(t, tables)
}
