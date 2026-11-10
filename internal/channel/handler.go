package channel

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// Handler expone el transporte HTTP del dominio de canales. Los errores del
// servicio se delegan a writeErr, el mapeo uniforme inyectado desde cmd/api.
type Handler struct {
	svc      *Service
	writeErr func(http.ResponseWriter, error)
}

// NewHandler crea el handler de canales con el servicio y el mapeo de
// errores inyectados.
func NewHandler(svc *Service, writeErr func(http.ResponseWriter, error)) *Handler {
	return &Handler{svc: svc, writeErr: writeErr}
}

type listResponse struct {
	Channels []Channel `json:"channels"`
}

// List responde GET /v1/channels con 200 {"channels": [...]}.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	channels, err := h.svc.List(r.Context())
	if err != nil {
		h.writeErr(w, err)
		return
	}
	if channels == nil {
		channels = []Channel{}
	}

	writeJSON(w, http.StatusOK, listResponse{Channels: channels})
}

// Create responde POST /v1/channels con 201 y el canal creado.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var ch Channel
	if err := json.NewDecoder(r.Body).Decode(&ch); err != nil {
		h.writeErr(w, ErrValidation{Field: "body", Msg: "must be a valid JSON channel"})
		return
	}

	created, err := h.svc.Create(r.Context(), ch)
	if err != nil {
		h.writeErr(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

// Get responde GET /v1/channels/{id} con 200 y el canal.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseID(w, r)
	if !ok {
		return
	}

	ch, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.writeErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, ch)
}

// Update responde PUT /v1/channels/{id} con 200 y el canal reemplazado.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseID(w, r)
	if !ok {
		return
	}

	var ch Channel
	if err := json.NewDecoder(r.Body).Decode(&ch); err != nil {
		h.writeErr(w, ErrValidation{Field: "body", Msg: "must be a valid JSON channel"})
		return
	}
	ch.ID = id

	updated, err := h.svc.Update(r.Context(), ch)
	if err != nil {
		h.writeErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

// Delete responde DELETE /v1/channels/{id} con 204 sin body.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseID(w, r)
	if !ok {
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.writeErr(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) parseID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		h.writeErr(w, ErrValidation{Field: "id", Msg: "must be a valid UUID"})
		return uuid.Nil, false
	}

	return id, true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
