package channel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type PostgresRepository struct {
	queries *Queries
	now     func() time.Time
}

func NewPostgresRepository(db DBTX) *PostgresRepository {
	return &PostgresRepository{
		queries: New(db),
		now:     func() time.Time { return time.Now().UTC() },
	}
}

func (r *PostgresRepository) List(ctx context.Context) ([]Channel, error) {
	rows, err := r.queries.ListChannels(ctx)
	if err != nil {
		return nil, err
	}

	channels := make([]Channel, len(rows))
	for i, row := range rows {
		channels[i], err = channelFromRow(row)
		if err != nil {
			return nil, err
		}
	}

	return channels, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (Channel, error) {
	row, err := r.queries.GetChannelByID(ctx, id)
	if err != nil {
		return Channel{}, mapPostgresError(err)
	}

	return channelFromRow(row)
}

func (r *PostgresRepository) Create(ctx context.Context, ch Channel) (Channel, error) {
	sources, err := json.Marshal(ch.Sources)
	if err != nil {
		return Channel{}, err
	}

	now := r.now().UTC()
	ch.ID = uuid.New()
	ch.CreatedAt = now
	ch.UpdatedAt = now

	row, err := r.queries.CreateChannel(ctx, CreateChannelParams{
		ID:             ch.ID,
		Number:         int32(ch.Number),
		Name:           ch.Name,
		Logo:           nullableText(ch.Logo),
		ScheduleUrl:    nullableText(ch.ScheduleURL),
		Enabled:        ch.Enabled,
		SegmentSeconds: int32(ch.SegmentSeconds),
		Sources:        sources,
		CreatedAt:      pgtype.Timestamptz{Time: ch.CreatedAt, Valid: true},
		UpdatedAt:      pgtype.Timestamptz{Time: ch.UpdatedAt, Valid: true},
	})
	if err != nil {
		return Channel{}, mapPostgresError(err)
	}

	return channelFromRow(row)
}

func (r *PostgresRepository) Update(ctx context.Context, ch Channel) (Channel, error) {
	sources, err := json.Marshal(ch.Sources)
	if err != nil {
		return Channel{}, err
	}

	ch.UpdatedAt = r.now().UTC()
	row, err := r.queries.UpdateChannel(ctx, UpdateChannelParams{
		ID:             ch.ID,
		Number:         int32(ch.Number),
		Name:           ch.Name,
		Logo:           nullableText(ch.Logo),
		ScheduleUrl:    nullableText(ch.ScheduleURL),
		Enabled:        ch.Enabled,
		SegmentSeconds: int32(ch.SegmentSeconds),
		Sources:        sources,
		UpdatedAt:      pgtype.Timestamptz{Time: ch.UpdatedAt, Valid: true},
	})
	if err != nil {
		return Channel{}, mapPostgresError(err)
	}

	return channelFromRow(row)
}

func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	rows, err := r.queries.DeleteChannel(ctx, id)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *PostgresRepository) ExistsNumber(ctx context.Context, number int, excludeID uuid.UUID) (bool, error) {
	return r.queries.ExistsChannelNumber(ctx, ExistsChannelNumberParams{
		Number: int32(number),
		ID:     excludeID,
	})
}

func channelFromRow(row ChannelRow) (Channel, error) {
	var ch Channel
	if err := json.Unmarshal(row.Sources, &ch.Sources); err != nil {
		return Channel{}, err
	}

	ch.ID = row.ID
	ch.Number = int(row.Number)
	ch.Name = row.Name
	ch.Logo = row.Logo.String
	ch.ScheduleURL = row.ScheduleUrl.String
	ch.Enabled = row.Enabled
	ch.SegmentSeconds = int(row.SegmentSeconds)
	ch.CreatedAt = row.CreatedAt.Time
	ch.UpdatedAt = row.UpdatedAt.Time

	return ch, nil
}

func nullableText(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: value != ""}
}

func mapPostgresError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: %v", ErrNotFound, err)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "channels_number_key" {
		return fmt.Errorf("%w: number already in use", ErrConflict)
	}

	return err
}

var _ Repository = (*PostgresRepository)(nil)
