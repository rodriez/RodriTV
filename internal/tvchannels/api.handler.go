package tvchannels

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	service TvChannelServiceGateway
}

func NewHandler(service TvChannelServiceGateway) Handler {
	return Handler{
		service: service,
	}
}

func (h *Handler) AddChannel(w http.ResponseWriter, r *http.Request) {
	var req TvChannelRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiErr := ApiErrBadRequest("invalid or malformed json", err.Error())
		apiErr.write(w)
		return
	}

	createdChannel, err := h.service.AddChannel(r.Context(), req.ToChannel())
	if err != nil {
		apiErr := FromError(err)
		apiErr.write(w)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Add("Content-Type", "applicaiton/json")
	json.NewEncoder(w).Encode(toResponse(&createdChannel))
}

func (h *Handler) GetChannelByID(w http.ResponseWriter, r *http.Request) {
	channelID := r.PathValue("channel_id")

	existingChannel, err := h.service.GetChannel(r.Context(), channelID)
	if err != nil {
		apiErr := FromError(err)
		apiErr.write(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "applicaiton/json")
	json.NewEncoder(w).Encode(toResponse(&existingChannel))
}

func (h *Handler) ReplaceChannel(w http.ResponseWriter, r *http.Request) {
	var req TvChannelRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiErr := ApiErrBadRequest("invalid or malformed json", err.Error())
		apiErr.write(w)
		return
	}

	replacedChannel, err := h.service.ReplaceChannel(r.Context(), req.ToChannel())
	if err != nil {
		apiErr := FromError(err)
		apiErr.write(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "applicaiton/json")
	json.NewEncoder(w).Encode(toResponse(&replacedChannel))
}

func (h *Handler) DeleteChannel(w http.ResponseWriter, r *http.Request) {
	channelID := r.PathValue("channel_id")

	deletedChannel, err := h.service.DeleteChannel(r.Context(), channelID)
	if err != nil {
		apiErr := FromError(err)
		apiErr.write(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "applicaiton/json")
	json.NewEncoder(w).Encode(toResponse(&deletedChannel))
}

func (h *Handler) ListChannels(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.ListChannels(r.Context())
	if err != nil {
		apiErr := FromError(err)
		apiErr.write(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "applicaiton/json")
	json.NewEncoder(w).Encode(toListResponse(list))
}
