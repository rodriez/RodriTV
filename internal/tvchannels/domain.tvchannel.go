package tvchannels

import (
	"slices"
	"time"
)

type TVChannelStatus string

const (
	TVChannelStatusActive  TVChannelStatus = "ACTIVE"
	TVChannelStatusPaused  TVChannelStatus = "PAUSED"
	TVChannelStatusDeleted TVChannelStatus = "DELETED"
)

var TvChannelValidStatuses []TVChannelStatus = []TVChannelStatus{
	TVChannelStatusActive,
	TVChannelStatusPaused,
	TVChannelStatusDeleted,
}

type TvChannel struct {
	ID           string
	Name         string
	ListPosition int
	Status       TVChannelStatus
	Signals      TvSignalSet
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (c *TvChannel) Empty() bool {
	return c.ID == "" &&
		c.Name == "" &&
		c.ListPosition == 0 &&
		c.Status == "" &&
		c.Signals.Empty() &&
		c.CreatedAt.IsZero() &&
		c.UpdatedAt.IsZero()
}

func (c *TvChannel) Validate() error {
	if c.Name == "" {
		return ErrInvalidName
	}

	if c.ListPosition < 0 {
		return ErrInvalidListPosition
	}

	if !slices.Contains(TvChannelValidStatuses, c.Status) {
		return ErrInvalidStatus
	}

	if c.Signals.Empty() {
		return ErrEmptySignals
	}

	if err := c.Signals.Validate(); err != nil {
		return err
	}

	return nil
}
