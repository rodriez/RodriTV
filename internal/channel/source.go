package channel

import (
	"time"

	"github.com/google/uuid"
)

// Source es una fuente de stream asignada a un slot (1-5) del canal.
type Source struct {
	ID        uuid.UUID `json:"id"`
	Slot      int       `json:"slot"`
	URL       string    `json:"url"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate verifica los invariantes de la source: slot dentro de 1-5
// y URL http(s) parseable.
func (s Source) Validate() error {
	if s.Slot < 1 || s.Slot > SourceSlots {
		return ErrValidation{Field: "slot", Msg: "must be between 1 and 5"}
	}
	if !isHTTPURL(s.URL) {
		return ErrValidation{Field: "url", Msg: "must be a valid http(s) URL"}
	}

	return nil
}
