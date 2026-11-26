package tvchannels

import "context"

//go:generate mockgen -source=domain.gateways.go -destination=domain.gateways_test.go -package=tvchannels_test
type TvChannelStoreGateway interface {
	Create(ctx context.Context, candidate *TvChannel) error
	GetByID(ctx context.Context, id string) (TvChannel, error)
	Replace(ctx context.Context, candidate *TvChannel) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]TvChannel, error)
}
