package contracts

import "time"

type TicketCheckInEmailData struct {
	FirstName  string
	LastName   *string
	Email      string
	TicketName string
	RegRef     string
	CheckedAt  *time.Time
}

type MailService interface {
	SendTicketCheckInEmail(data TicketCheckInEmailData) error
}
