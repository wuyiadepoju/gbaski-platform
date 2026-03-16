package contracts

import (
	"context"

	"github.com/gbaski/gbaski-platform/internal/form/domain"
	"github.com/google/uuid"
)

// FormRepository defines the persistence operations for forms
type FormRepository interface {
	CreateForm(ctx context.Context, form *domain.FormAggregate) (*uuid.UUID, error)
	UpdateForm(ctx context.Context, form *domain.FormAggregate) error
	GetFormsByEventId(ctx context.Context, eventId uuid.UUID) ([]domain.FormQueryModel, error)
}

