package channel

import (
	"context"

	"github.com/google/uuid"
)

// Repository es la interfaz de salida del dominio channel hacia la
// persistencia, definida por el consumidor (el servicio de canales).
//
// Contratos de error:
//   - GetByID, Update y Delete devuelven ErrNotFound cuando el canal no existe.
//   - Create y Update devuelven ErrConflict cuando el number ya está en uso.
type Repository interface {
	// List devuelve todos los canales ordenados por number ascendente.
	List(ctx context.Context) ([]Channel, error)

	// GetByID devuelve el canal con el id dado o ErrNotFound si no existe.
	GetByID(ctx context.Context, id uuid.UUID) (Channel, error)

	// Create persiste un canal nuevo y devuelve la versión persistida.
	// Devuelve ErrConflict si el number ya está en uso.
	Create(ctx context.Context, ch Channel) (Channel, error)

	// Update persiste los cambios de un canal existente identificado por ch.ID
	// y devuelve la versión persistida. Devuelve ErrNotFound si no existe y
	// ErrConflict si el number ya está en uso por otro canal.
	Update(ctx context.Context, ch Channel) (Channel, error)

	// Delete elimina el canal con el id dado. Devuelve ErrNotFound si no existe.
	Delete(ctx context.Context, id uuid.UUID) error

	// ExistsNumber indica si ya existe un canal con ese number, excluyendo el
	// canal excludeID (para validar conflicto 409 en actualizaciones).
	ExistsNumber(ctx context.Context, number int, excludeID uuid.UUID) (bool, error)
}
