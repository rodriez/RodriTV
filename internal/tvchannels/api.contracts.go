package tvchannels

import "time"

type TvChannelRequest struct {
	ID           string          `json:"id,omitempty"`
	Name         string          `json:"name,omitempty"`
	ListPosition int             `json:"listPosition,omitempty"`
	Status       TVChannelStatus `json:"status,omitempty"`
	Signals      []struct {
		URL         string `json:"url,omitempty"`
		Priority    int    `json:"priority,omitempty"`
		NeedBrowser bool   `json:"needBrowser,omitempty"`
	} `json:"signals"`
}

func (req *TvChannelRequest) ToChannel() TvChannel {
	signals := TvSignalSet{}
	for i, sig := range req.Signals {
		signals[i].URL = sig.URL
		signals[i].Priority = sig.Priority
		signals[i].NeedBrowser = sig.NeedBrowser
	}

	return TvChannel{
		ID:           req.ID,
		Name:         req.Name,
		ListPosition: req.ListPosition,
		Status:       req.Status,
		Signals:      signals,
	}
}

type TvChannelResponse struct {
	ID           string          `json:"id,omitempty"`
	Name         string          `json:"name,omitempty"`
	ListPosition int             `json:"listPosition,omitempty"`
	Status       TVChannelStatus `json:"status,omitempty"`
	Signals      []struct {
		URL         string `json:"url,omitempty"`
		Priority    int    `json:"priority,omitempty"`
		NeedBrowser bool   `json:"needBrowser,omitempty"`
	} `json:"signals"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

func toResponse(tvchannel *TvChannel) TvChannelResponse {
	signals := []struct {
		URL         string `json:"url,omitempty"`
		Priority    int    `json:"priority,omitempty"`
		NeedBrowser bool   `json:"needBrowser,omitempty"`
	}{}

	for _, sig := range tvchannel.Signals {
		if sig.Empty() {
			continue
		}

		signals = append(signals, struct {
			URL         string `json:"url,omitempty"`
			Priority    int    `json:"priority,omitempty"`
			NeedBrowser bool   `json:"needBrowser,omitempty"`
		}{
			URL:         sig.URL,
			Priority:    sig.Priority,
			NeedBrowser: sig.NeedBrowser,
		})
	}

	return TvChannelResponse{
		ID:           tvchannel.ID,
		Name:         tvchannel.Name,
		ListPosition: tvchannel.ListPosition,
		Status:       tvchannel.Status,
		Signals:      signals,
		CreatedAt:    tvchannel.CreatedAt,
		UpdatedAt:    tvchannel.UpdatedAt,
	}
}

func toListResponse(channels []TvChannel) []TvChannelResponse {
	list := make([]TvChannelResponse, len(channels))
	for i, channel := range channels {
		list[i] = toResponse(&channel)
	}

	return list
}
