package tvchannels

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rodriez/RodriTV/internal/sqlc"
)

type TvChannelsStoreRepository struct {
	db *sqlc.Queries
}

func NewRepository(db *sqlc.Queries) TvChannelsStoreRepository {
	return TvChannelsStoreRepository{
		db: db,
	}
}

func (repo *TvChannelsStoreRepository) Create(ctx context.Context, candidate *TvChannel) error {
	signalsJson, err := json.Marshal(candidate.Signals)
	if err != nil {
		slog.Error("error marshalling signals to json", "error", err.Error())
		return ErrInternal
	}

	args := sqlc.CreateChannelParams{
		ID:           candidate.ID,
		Name:         candidate.Name,
		ListPosition: int32(candidate.ListPosition),
		Status:       string(candidate.Status),
		Signals:      signalsJson,
		CreatedAt: pgtype.Timestamptz{
			Time:  candidate.CreatedAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  candidate.UpdatedAt,
			Valid: true,
		},
	}

	_, err = repo.db.CreateChannel(ctx, args)
	if err != nil {
		slog.Error("error trying to insert tvchannel", "args", args, "error", err.Error())
		return ErrDBInternal
	}

	return err
}

func (repo *TvChannelsStoreRepository) GetByID(ctx context.Context, id string) (TvChannel, error) {
	resultSet, err := repo.db.GetChannelByID(ctx, id)
	if err != nil {
		slog.Error("error trying to insert tvchannel", "id", id, "error", err.Error())
		return TvChannel{}, ErrDBInternal
	}

	channel, err := Record(resultSet).ToTvChannel()
	if err != nil {
		return TvChannel{}, err
	}

	return channel, nil
}

func (repo *TvChannelsStoreRepository) Replace(ctx context.Context, candidate *TvChannel) error {
	signalsJson, err := json.Marshal(candidate.Signals)
	if err != nil {
		slog.Error("error marshalling signals to json", "error", err.Error())
		return ErrInternal
	}

	args := sqlc.ReplaceChannelParams{
		ID:           candidate.ID,
		Name:         candidate.Name,
		ListPosition: int32(candidate.ListPosition),
		Status:       string(candidate.Status),
		Signals:      signalsJson,
		UpdatedAt: pgtype.Timestamptz{
			Time:  candidate.UpdatedAt,
			Valid: true,
		},
	}

	_, err = repo.db.ReplaceChannel(ctx, args)
	if err != nil {
		slog.Error("error trying to insert tvchannel", "id", candidate.ID, "args", args, "error", err.Error())
		return ErrDBInternal
	}

	return err
}

func (repo *TvChannelsStoreRepository) Delete(ctx context.Context, id string) error {
	err := repo.db.DeleteChannel(ctx, id)
	if err != nil {
		slog.Error("error trying to delete tvchannel", "id", id, "error", err.Error())
		return ErrDBInternal
	}

	return nil
}

func (repo *TvChannelsStoreRepository) List(ctx context.Context) ([]TvChannel, error) {
	resultSet, err := repo.db.ListChannels(ctx)
	if err != nil {
		slog.Error("error trying to get tvchannels list", "error", err.Error())
		return nil, ErrInternal
	}

	list := []TvChannel{}

	for _, rs := range resultSet {
		channel, err := Record(rs).ToTvChannel()
		if err != nil {
			slog.Warn("ingoring tv channel from list", "id", rs.ID, "error", err.Error())
			continue
		}

		list = append(list, channel)
	}

	return list, nil
}
