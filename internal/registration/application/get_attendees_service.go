package application

import (
	"context"

	"github.com/gbaski/gbaski-platform/internal/registration/contracts"
	"github.com/gbaski/gbaski-platform/internal/registration/domain"
)

type GetAttendeesService struct {
	repo contracts.RegistrationRepository
}

func NewGetAttendeesService(repo contracts.RegistrationRepository) *GetAttendeesService {
	return &GetAttendeesService{repo: repo}
}

func (s *GetAttendeesService) Execute(ctx context.Context, query GetAttendeesQuery) (domain.TicketAttendeeList, error) {
	return s.repo.GetTicketAttendees(ctx, query.Query, query.FormID, query.UserID)
}

type LookupRegistrationService struct {
	repo contracts.RegistrationRepository
}

func NewLookupRegistrationService(repo contracts.RegistrationRepository) *LookupRegistrationService {
	return &LookupRegistrationService{repo: repo}
}

func (s *LookupRegistrationService) Execute(ctx context.Context, query LookupRegistrationQuery) (*AttendeeDTO, error) {
	attendee, err := s.repo.LookupTicketRegistration(ctx, query.FormID, query.RegRef)
	if err != nil {
		return nil, err
	}
	dto := ToAttendeeDTO(*attendee)
	return &dto, nil
}
