package adapters

import (
	"fmt"

	"github.com/gbaski/gbaski-ext/mailcoach"
	"github.com/gbaski/gbaski-platform/internal/registration/contracts"
)

type MailServiceImpl struct {
	mailcoach *mailcoach.Mailcoach
}

func NewMailServiceImpl() contracts.MailService {
	return &MailServiceImpl{mailcoach: mailcoach.New()}
}

func (s *MailServiceImpl) SendTicketCheckInEmail(data contracts.TicketCheckInEmailData) error {
	// For now, let's use the simple transactional email or a dedicated template if available.
	// In the original code, this was calling eventService.SendTicketCheckInEmail.
	// We'll implement the actual sending logic here to keep registration independent of event service if possible,
	// or we can wrap the event service if it's considered an infrastructure dependency.
	// But let's try to be direct with mailcoach.

	subject := fmt.Sprintf("Checked-in: %s", data.TicketName)
	htmlContent := fmt.Sprintf("<p>Hi %s, you have been checked in for %s.</p>", data.FirstName, data.TicketName)

	return s.mailcoach.SendSimpleTransactionalEmail(data.Email, subject, htmlContent)
}
