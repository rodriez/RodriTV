package tvchannels_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rodriez/RodriTV/internal/sqlc"
	"github.com/rodriez/RodriTV/internal/tvchannels"
	"github.com/shibukawa/pgmem"
	"github.com/stretchr/testify/suite"
)

type RepositoryTestSuite struct {
	suite.Suite

	server *pgmem.Server
	pool   *pgxpool.Pool
}

// SetupSuite starts an in-memory PostgreSQL inside the test process.
func (s *RepositoryTestSuite) SetupSuite() {
	ctx := context.Background()

	server, err := pgmem.Start(ctx, pgmem.Options{Database: "rodritv"})
	s.Require().NoError(err)
	s.server = server

	s.pool, err = pgxpool.New(ctx, server.DSN())
	s.Require().NoError(err)

	schema, err := os.ReadFile("../../scripts/db/migrations/0001_channels.sql")
	s.Require().NoError(err)

	_, err = s.pool.Exec(ctx, string(schema))
	s.Require().NoError(err)
}

func (s *RepositoryTestSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}

	if s.server != nil {
		s.NoError(s.server.Close())
	}
}

func (s *RepositoryTestSuite) SetupTest() {
	_, err := s.pool.Exec(context.Background(), "TRUNCATE TABLE channels")
	s.Require().NoError(err)
}

func (s *RepositoryTestSuite) newTvChannel() tvchannels.TvChannel {
	// PostgreSQL keeps microseconds, so the fixture has no nanoseconds
	now := time.Date(2026, time.January, 15, 10, 30, 0, 123456000, time.UTC)

	return tvchannels.TvChannel{
		ID:           "3b871028-be80-4257-8759-53d19f7fbd92",
		Name:         "Telefe",
		ListPosition: 1,
		Status:       tvchannels.TVChannelStatusPaused,
		Signals: tvchannels.TvSignalSet{
			{
				URL:         "https://www.mitelefe.com/telefe-en-vivo",
				Priority:    1,
				NeedBrowser: true,
			},
			{
				URL:                "http://45.134.141.161:2200/ARG/TELEFE_HD/index.m3u8",
				Priority:           2,
				NeedRefresh:        true,
				RefreshRateSeconds: 30,
				BufferEnabled:      true,
				BufferSize:         1024,
			},
		},
		CreatedAt: now,
		UpdatedAt: now.Add(time.Minute),
	}
}

func (s *RepositoryTestSuite) repository() tvchannels.TvChannelsStoreRepository {
	return tvchannels.NewRepository(sqlc.New(s.pool))
}

func (s *RepositoryTestSuite) countChannels() int {
	var count int
	err := s.pool.QueryRow(context.Background(), "SELECT count(*) FROM channels").Scan(&count)
	s.Require().NoError(err)

	return count
}

func (s *RepositoryTestSuite) assertStored(expected tvchannels.TvChannel) {
	repo := s.repository()
	stored, err := repo.GetByID(context.Background(), expected.ID)
	s.Require().NoError(err)

	s.assertEqualChannel(expected, stored)
}

func (s *RepositoryTestSuite) assertEqualChannel(expected, stored tvchannels.TvChannel) {
	// Compare instants, the driver returns them in the local time zone
	s.True(expected.CreatedAt.Equal(stored.CreatedAt), "createdAt: %s", stored.CreatedAt)
	s.True(expected.UpdatedAt.Equal(stored.UpdatedAt), "updatedAt: %s", stored.UpdatedAt)

	stored.CreatedAt = expected.CreatedAt
	stored.UpdatedAt = expected.UpdatedAt
	s.Equal(expected, stored)
}

func (s *RepositoryTestSuite) TestCreateSuccessfully() {
	candidate := s.newTvChannel()
	repo := s.repository()

	err := repo.Create(context.Background(), &candidate)

	s.Require().NoError(err)
	s.Equal(1, s.countChannels())
	s.assertStored(candidate)
}

func (s *RepositoryTestSuite) TestCreateWithDuplicatedID() {
	original := s.newTvChannel()
	duplicated := s.newTvChannel()
	duplicated.Name = "Telefe Online"

	repo := s.repository()
	s.Require().NoError(repo.Create(context.Background(), &original))

	err := repo.Create(context.Background(), &duplicated)

	s.ErrorIs(err, tvchannels.ErrDBInternal)
	s.Equal(1, s.countChannels())
	s.assertStored(original)
}

func (s *RepositoryTestSuite) TestCreateWithCanceledContext() {
	candidate := s.newTvChannel()
	repo := s.repository()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := repo.Create(ctx, &candidate)

	s.ErrorIs(err, tvchannels.ErrDBInternal)
	s.Equal(0, s.countChannels())
}

func (s *RepositoryTestSuite) TestGetByIDSuccessfully() {
	existing := s.newTvChannel()
	other := s.newTvChannel()
	other.ID = "9f1c2d4e-5b6a-4c7d-8e9f-0a1b2c3d4e5f"
	other.Name = "El Trece"

	repo := s.repository()
	s.Require().NoError(repo.Create(context.Background(), &existing))
	s.Require().NoError(repo.Create(context.Background(), &other))

	// assertStored reads through GetByID and compares every field
	s.assertStored(existing)
	s.assertStored(other)
}

func (s *RepositoryTestSuite) TestGetByIDWithNonRegisteredID() {
	repo := s.repository()

	result, err := repo.GetByID(context.Background(), "3b871028-be80-4257-8759-53d19f7fbd92")

	s.NoError(err)
	s.True(result.Empty())
}

func (s *RepositoryTestSuite) TestGetByIDWithInvalidStoredJson() {
	id := "3b871028-be80-4257-8759-53d19f7fbd92"

	// Valid for the JSON column, but it is not a list of signals
	_, err := s.pool.Exec(context.Background(),
		`INSERT INTO channels (id, name, list_position, status, signals) VALUES ($1, 'Telefe', 1, 'PAUSED', '{}')`, id)
	s.Require().NoError(err)

	repo := s.repository()
	result, err := repo.GetByID(context.Background(), id)

	s.ErrorIs(err, tvchannels.ErrInvalidStoredJson)
	s.True(result.Empty())
}

func (s *RepositoryTestSuite) TestGetByIDWithCanceledContext() {
	repo := s.repository()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := repo.GetByID(ctx, "3b871028-be80-4257-8759-53d19f7fbd92")

	s.ErrorIs(err, tvchannels.ErrDBInternal)
	s.True(result.Empty())
}

func (s *RepositoryTestSuite) TestReplaceSuccessfully() {
	existing := s.newTvChannel()
	other := s.newTvChannel()
	other.ID = "9f1c2d4e-5b6a-4c7d-8e9f-0a1b2c3d4e5f"

	repo := s.repository()
	s.Require().NoError(repo.Create(context.Background(), &existing))
	s.Require().NoError(repo.Create(context.Background(), &other))

	candidate := existing
	candidate.Name = "Telefe Online"
	candidate.ListPosition = 7
	candidate.Status = tvchannels.TVChannelStatusActive
	candidate.Signals = tvchannels.TvSignalSet{{URL: "http://example.com/telefe.m3u8", Priority: 1}}
	candidate.CreatedAt = existing.CreatedAt.Add(time.Hour)
	candidate.UpdatedAt = existing.UpdatedAt.Add(time.Hour)

	err := repo.Replace(context.Background(), &candidate)

	s.Require().NoError(err)
	s.Equal(2, s.countChannels())

	// createdAt is never replaced
	expected := candidate
	expected.CreatedAt = existing.CreatedAt
	s.assertStored(expected)
	s.assertStored(other)
}

func (s *RepositoryTestSuite) TestReplaceWithNonRegisteredID() {
	candidate := s.newTvChannel()
	repo := s.repository()

	err := repo.Replace(context.Background(), &candidate)

	// Not an error for the repository, the service checks the ID exists
	s.NoError(err)
	s.Equal(0, s.countChannels())
}

func (s *RepositoryTestSuite) TestReplaceWithCanceledContext() {
	existing := s.newTvChannel()
	repo := s.repository()
	s.Require().NoError(repo.Create(context.Background(), &existing))

	candidate := existing
	candidate.Name = "Telefe Online"

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := repo.Replace(ctx, &candidate)

	s.ErrorIs(err, tvchannels.ErrDBInternal)
	// The replace failed, so the stored row must still be the existing one and
	// not the candidate. Both share the ID, only the content tells them apart
	s.assertStored(existing)
}

func (s *RepositoryTestSuite) TestDeleteSuccessfully() {
	existing := s.newTvChannel()
	other := s.newTvChannel()
	other.ID = "9f1c2d4e-5b6a-4c7d-8e9f-0a1b2c3d4e5f"

	repo := s.repository()
	s.Require().NoError(repo.Create(context.Background(), &existing))
	s.Require().NoError(repo.Create(context.Background(), &other))

	err := repo.Delete(context.Background(), existing.ID)

	s.Require().NoError(err)
	s.Equal(1, s.countChannels())
	// The only row left is the one that was not deleted
	s.assertStored(other)
}

func (s *RepositoryTestSuite) TestDeleteWithNonRegisteredID() {
	existing := s.newTvChannel()
	repo := s.repository()
	s.Require().NoError(repo.Create(context.Background(), &existing))

	err := repo.Delete(context.Background(), "9f1c2d4e-5b6a-4c7d-8e9f-0a1b2c3d4e5f")

	// Not an error for the repository, the service checks the ID exists
	s.NoError(err)
	s.Equal(1, s.countChannels())
	s.assertStored(existing)
}

func (s *RepositoryTestSuite) TestDeleteWithCanceledContext() {
	existing := s.newTvChannel()
	repo := s.repository()
	s.Require().NoError(repo.Create(context.Background(), &existing))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := repo.Delete(ctx, existing.ID)

	s.ErrorIs(err, tvchannels.ErrDBInternal)
	// The delete failed, so the existing row must still be stored
	s.assertStored(existing)
}

func (s *RepositoryTestSuite) TestListSuccessfully() {
	first := s.newTvChannel()
	second := s.newTvChannel()
	second.ID = "9f1c2d4e-5b6a-4c7d-8e9f-0a1b2c3d4e5f"
	second.Name = "El Trece"
	second.ListPosition = 2

	// Inserted in reverse order, to check the list is sorted by listPosition
	repo := s.repository()
	s.Require().NoError(repo.Create(context.Background(), &second))
	s.Require().NoError(repo.Create(context.Background(), &first))

	list, err := repo.List(context.Background())

	s.Require().NoError(err)
	s.Require().Len(list, 2)
	s.assertEqualChannel(first, list[0])
	s.assertEqualChannel(second, list[1])
}

func (s *RepositoryTestSuite) TestListWithoutChannels() {
	repo := s.repository()

	list, err := repo.List(context.Background())

	s.NoError(err)
	s.NotNil(list)
	s.Empty(list)
}

func (s *RepositoryTestSuite) TestListWithInvalidStoredJson() {
	valid := s.newTvChannel()
	repo := s.repository()
	s.Require().NoError(repo.Create(context.Background(), &valid))

	// Valid for the JSON column, but it is not a list of signals
	_, err := s.pool.Exec(context.Background(),
		`INSERT INTO channels (id, name, list_position, status, signals) VALUES ('9f1c2d4e-5b6a-4c7d-8e9f-0a1b2c3d4e5f', 'El Trece', 2, 'PAUSED', '{}')`)
	s.Require().NoError(err)

	list, err := repo.List(context.Background())

	// The broken row is ignored, it does not fail the whole list
	s.Require().NoError(err)
	s.Require().Len(list, 1)
	s.assertEqualChannel(valid, list[0])
}

func (s *RepositoryTestSuite) TestListWithCanceledContext() {
	repo := s.repository()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	list, err := repo.List(ctx)

	// Unlike the other methods, List returns ErrInternal
	s.ErrorIs(err, tvchannels.ErrInternal)
	s.Nil(list)
}

func TestRepository(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}
