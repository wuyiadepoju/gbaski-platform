package contracts

import (
	"context"

	"github.com/google/uuid"
)

type CheckInAgentStore interface {
	SetAgent(ctx context.Context, formID uuid.UUID, id uuid.UUID, name string) error
	GetAgent(ctx context.Context, formID uuid.UUID, id uuid.UUID) (string, error)
	DeleteAgent(ctx context.Context, formID uuid.UUID, id uuid.UUID) error
	GetAllAgents(ctx context.Context, formID uuid.UUID) (map[string]string, error)
}
