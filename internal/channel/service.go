package channel

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// LogoStore elimina los archivos de logo asociados a canales.
type LogoStore interface {
	Delete(ctx context.Context, name string) error
}

// Service orquesta los casos de uso del dominio de canales.
type Service struct {
	repo      Repository
	now       func() time.Time
	logoStore LogoStore
}

// NewService crea un servicio de canales con reloj inyectable.
func NewService(repo Repository, now func() time.Time) *Service {
	return &Service{repo: repo, now: now}
}

// SetLogoStore configura el almacenamiento de logos usado al eliminar canales.
func (s *Service) SetLogoStore(store LogoStore) {
	s.logoStore = store
}

func (s *Service) List(ctx context.Context) ([]Channel, error) {
	return s.repo.List(ctx)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (Channel, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, ch Channel) (Channel, error) {
	if ch.SegmentSeconds == 0 {
		ch.SegmentSeconds = DefaultSegmentSeconds
	}
	ch.Enabled = true
	s.normalizeSources(&ch)

	if err := ch.Validate(); err != nil {
		return Channel{}, err
	}
	exists, err := s.repo.ExistsNumber(ctx, ch.Number, uuid.Nil)
	if err != nil {
		return Channel{}, err
	}
	if exists {
		return Channel{}, ErrConflict
	}

	return s.repo.Create(ctx, ch)
}

func (s *Service) Update(ctx context.Context, ch Channel) (Channel, error) {
	s.normalizeSources(&ch)
	if err := ch.Validate(); err != nil {
		return Channel{}, err
	}
	exists, err := s.repo.ExistsNumber(ctx, ch.Number, ch.ID)
	if err != nil {
		return Channel{}, err
	}
	if exists {
		return Channel{}, ErrConflict
	}

	return s.repo.Update(ctx, ch)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	ch, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	if ch.Logo != "" && s.logoStore != nil {
		return s.logoStore.Delete(ctx, ch.Logo)
	}

	return nil
}

func (s *Service) normalizeSources(ch *Channel) {
	now := s.now()
	for _, source := range ch.Sources {
		if source == nil {
			continue
		}
		if source.ID == uuid.Nil {
			source.ID = uuid.New()
			source.CreatedAt = now
		}
		source.UpdatedAt = now
	}
}
