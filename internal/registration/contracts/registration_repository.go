package contracts

import (
	"context"

	"github.com/gbaski/gbaski-platform/internal/registration/domain"
	"github.com/google/uuid"
)

type RegistrationRepository interface {
	GetRegistrations(ctx context.Context, query domain.RegistrationQueryModel, formID uuid.UUID, regType string, userID uuid.UUID) (domain.RegistrationList, error)
	GetRegistrationStats(ctx context.Context, formID uuid.UUID, regType string, userID uuid.UUID) (domain.RegistrationStats, error)
	GetTicketAttendees(ctx context.Context, query domain.RegistrationQueryModel, formID uuid.UUID, userID uuid.UUID) (domain.TicketAttendeeList, error)
	LookupTicketRegistration(ctx context.Context, formID uuid.UUID, regRef string) (*domain.Attendee, error)
	SaveAttendee(ctx context.Context, attendee *domain.Attendee) error
}
