// Package channel modela el dominio de canales: entidades, invariantes
// y errores de dominio, sin dependencias de infraestructura.
package channel

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
)

// Errores sentinelas de dominio, mapeados a HTTP en cmd/api/server.go.
var (
	ErrNotFound = errors.New("channel not found")
	ErrConflict = errors.New("channel conflict")
)

// ErrValidation es el error tipado de validación de invariantes de dominio.
type ErrValidation struct {
	Field string
	Msg   string
}

func (e ErrValidation) Error() string {
	return fmt.Sprintf("invalid %s: %s", e.Field, e.Msg)
}

// ValidationFields expone los campos inválidos para el mapeo HTTP uniforme
// (interface validationError de cmd/api/server.go).
func (e ErrValidation) ValidationFields() map[string]string {
	return map[string]string{e.Field: e.Msg}
}

const (
	// SourceSlots es la cantidad fija de posiciones de sources por canal.
	SourceSlots = 5
	// DefaultSegmentSeconds es el tamaño de segmento HLS por defecto.
	DefaultSegmentSeconds = 5

	maxNameLen        = 100
	minSegmentSeconds = 2
	maxSegmentSeconds = 10
)

// Channel es la entidad de dominio de un canal con sus 5 slots de sources.
type Channel struct {
	ID             uuid.UUID            `json:"id"`
	Number         int                  `json:"number"`
	Name           string               `json:"name"`
	Logo           string               `json:"logo,omitempty"`
	ScheduleURL    string               `json:"schedule_url,omitempty"`
	Enabled        bool                 `json:"enabled"`
	SegmentSeconds int                  `json:"segment_seconds"`
	Sources        [SourceSlots]*Source `json:"sources"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

// NewChannel crea un canal con los defaults de dominio: habilitado,
// SegmentSeconds por defecto y timestamps de creación.
func NewChannel(number int, name string) Channel {
	now := time.Now()

	return Channel{
		ID:             uuid.New(),
		Number:         number,
		Name:           name,
		Enabled:        true,
		SegmentSeconds: DefaultSegmentSeconds,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// Validate verifica los invariantes del canal y de sus sources ocupados.
func (c Channel) Validate() error {
	if c.Number <= 0 {
		return ErrValidation{Field: "number", Msg: "must be a positive integer"}
	}
	if len(c.Name) == 0 || len(c.Name) > maxNameLen {
		return ErrValidation{Field: "name", Msg: "must be between 1 and 100 characters"}
	}
	if c.SegmentSeconds < minSegmentSeconds || c.SegmentSeconds > maxSegmentSeconds {
		return ErrValidation{Field: "segment_seconds", Msg: "must be between 2 and 10"}
	}
	if c.ScheduleURL != "" && !isHTTPURL(c.ScheduleURL) {
		return ErrValidation{Field: "schedule_url", Msg: "must be a valid http(s) URL"}
	}
	for i, s := range c.Sources {
		if s == nil {
			continue
		}
		if err := s.Validate(); err != nil {
			return err
		}
		if s.Slot != i+1 {
			return ErrValidation{Field: "sources", Msg: fmt.Sprintf("source in position %d must have slot %d", i+1, i+1)}
		}
	}

	return nil
}

func isHTTPURL(raw string) bool {
	u, err := url.Parse(raw)

	return err == nil && u.Host != "" && (u.Scheme == "http" || u.Scheme == "https")
}
