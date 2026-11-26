package tvchannels

import (
	"context"
	"log/slog"
	"time"
	"uuid"
)

type Service struct {
	store TvChannelStoreGateway
}

func NewService(store TvChannelStoreGateway) Service {
	return Service{
		store: store,
	}
}

func (srv *Service) AddChannel(ctx context.Context, candidate TvChannel) (TvChannel, error) {
	candidate.ID = uuid.NewV4().String()
	candidate.Status = TVChannelStatusPaused
	candidate.CreatedAt = time.Now()
	candidate.UpdatedAt = time.Now()

	if err := candidate.Validate(); err != nil {
		return TvChannel{}, err
	}

	if err := srv.store.Create(ctx, &candidate); err != nil {
		slog.Error("error ocurred trying to create channel", "error", err)
		return TvChannel{}, ErrInternal
	}

	return candidate, nil
}

func (srv *Service) GetChannel(ctx context.Context, id string) (TvChannel, error) {
	if id == "" {
		return TvChannel{}, ErrInvalidID
	}

	tvchannel, err := srv.store.GetByID(ctx, id)

	if err != nil {
		slog.Error("error ocurred trying to get channel", "error", err)
		return TvChannel{}, ErrInternal
	}

	if tvchannel.Empty() {
		return TvChannel{}, ErrInvalidID
	}

	return tvchannel, nil
}

func (srv *Service) ReplaceChannel(ctx context.Context, candidate TvChannel) (TvChannel, error) {
	existing, err := srv.GetChannel(ctx, candidate.ID)
	if err != nil {
		return TvChannel{}, err
	}

	candidate.UpdatedAt = time.Now()
	candidate.CreatedAt = existing.CreatedAt

	if err := candidate.Validate(); err != nil {
		return TvChannel{}, err
	}

	if err := srv.store.Replace(ctx, &candidate); err != nil {
		slog.Error("error ocurred trying to replace channel", "error", err)
		return TvChannel{}, ErrInternal
	}

	return candidate, nil
}

func (srv *Service) DeleteChannel(ctx context.Context, id string) (TvChannel, error) {
	existing, err := srv.GetChannel(ctx, id)
	if err != nil {
		return TvChannel{}, err
	}

	if err := srv.store.Delete(ctx, id); err != nil {
		slog.Error("error ocurred trying to delete channel", "error", err)
		return TvChannel{}, ErrInternal
	}

	return existing, nil
}

func (srv *Service) ListChannels(ctx context.Context) ([]TvChannel, error) {
	list, err := srv.store.List(ctx)

	if err != nil {
		slog.Error("error ocurred trying to get channel", "error", err)
		return nil, ErrInternal
	}

	return list, nil
}
