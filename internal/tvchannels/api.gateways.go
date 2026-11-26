package tvchannels

import "context"

//go:generate mockgen -source=api.gateways.go -destination=api.gateways_test.go -package=tvchannels_test

type TvChannelServiceGateway interface {
	AddChannel(ctx context.Context, tvChannel TvChannel) (TvChannel, error)
	GetChannel(ctx context.Context, id string) (TvChannel, error)
	ReplaceChannel(ctx context.Context, candidate TvChannel) (TvChannel, error)
	DeleteChannel(ctx context.Context, id string) (TvChannel, error)
	ListChannels(ctx context.Context) ([]TvChannel, error)
}
