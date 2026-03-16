package application

import (
	"context"

	"github.com/gbaski/gbaski-ext/log"
	"github.com/gbaski/gbaski-platform/internal/registration/contracts"
	"github.com/google/uuid"
)

type CheckInService struct {
	repo        contracts.RegistrationRepository
	mailService contracts.MailService
}

func NewCheckInService(repo contracts.RegistrationRepository, mailService contracts.MailService) *CheckInService {
	return &CheckInService{
		repo:        repo,
		mailService: mailService,
	}
}

func (s *CheckInService) Execute(ctx context.Context, cmd CheckInCommand) error {
	// 1. Lookup registration by ref
	attendee, err := s.repo.LookupTicketRegistration(ctx, uuid.Nil, cmd.RegRef)
	if err != nil {
		return err
	}

	// 2. Perform check-in logic on the domain aggregate
	if err := attendee.CheckIn(cmd.AgentName); err != nil {
		return err
	}

	// 3. Save the attendee
	if err := s.repo.SaveAttendee(ctx, attendee); err != nil {
		return err
	}

	// 4. Send email asynchronously
	go func() {
		emailData := contracts.TicketCheckInEmailData{
			FirstName:  attendee.FirstName,
			LastName:   attendee.LastName,
			Email:      attendee.Email,
			TicketName: attendee.TicketName,
			RegRef:     attendee.RegRef,
			CheckedAt:  attendee.CheckedAt,
		}

		err := s.mailService.SendTicketCheckInEmail(emailData)
		if err != nil {
			log.Error("registration", "check_in_service", err)
		}
	}()

	return nil
}
