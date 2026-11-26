package tvchannels

import (
	"encoding/json"

	"github.com/rodriez/RodriTV/internal/sqlc"
)

type Record sqlc.Channel

func (r Record) ToTvChannel() (TvChannel, error) {
	var signals TvSignalSet

	err := json.Unmarshal(r.Signals, &signals)
	if err != nil {
		return TvChannel{}, ErrInvalidStoredJson
	}

	channel := TvChannel{
		ID:           r.ID,
		Name:         r.Name,
		ListPosition: int(r.ListPosition),
		Status:       TVChannelStatus(r.Status),
		Signals:      signals,
		CreatedAt:    r.CreatedAt.Time,
		UpdatedAt:    r.UpdatedAt.Time,
	}

	return channel, nil
}
