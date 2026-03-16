package application

import (
	"github.com/gbaski/gbaski-platform/internal/registration/domain"
	"github.com/google/uuid"
)

type GetRegistrationsQuery struct {
	Query  domain.RegistrationQueryModel
	FormID uuid.UUID
	UserID uuid.UUID
}

type GetAttendeesQuery struct {
	Query  domain.RegistrationQueryModel
	FormID uuid.UUID
	UserID uuid.UUID
}

type GetRegistrationStatsQuery struct {
	FormID uuid.UUID
	UserID uuid.UUID
}

type CheckInCommand struct {
	RegRef    string
	AgentName string
}

type GetAgentsQuery struct {
	FormID uuid.UUID
}

type CreateAgentCommand struct {
	FormID uuid.UUID
	Name   string
}

type DeleteAgentCommand struct {
	FormID uuid.UUID
	Token  string
}

type GetAgentByTokenQuery struct {
	Token string
}

type LookupRegistrationQuery struct {
	FormID uuid.UUID
	RegRef string
}
