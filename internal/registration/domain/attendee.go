package domain

import (
	"time"
)

type Attendee struct {
	ID         int64
	RegID      int
	FirstName  string
	LastName   *string
	Email      string
	Phone      string
	RegRef     string
	TicketID   int64
	TicketName string
	CheckedIn  bool
	CheckedAt  *time.Time
	CheckedBy  *string
	CreatedAt  time.Time

	// Domain events
	domainEvents []DomainEvent
}

func NewAttendee(regID int, firstName, email, phone, regRef string, ticketID int64) (*Attendee, error) {
	if firstName == "" {
		return nil, ErrEmptyFirstName
	}
	if email == "" {
		return nil, ErrEmptyEmail
	}

	return &Attendee{
		RegID:        regID,
		FirstName:    firstName,
		Email:        email,
		Phone:        phone,
		RegRef:       regRef,
		TicketID:     ticketID,
		CheckedIn:    false,
		CreatedAt:    time.Now(),
		domainEvents: []DomainEvent{},
	}, nil
}

func (a *Attendee) CheckIn(agentName string) error {
	if a.CheckedIn {
		return ErrAlreadyCheckedIn
	}

	now := time.Now()
	a.CheckedIn = true
	a.CheckedAt = &now
	a.CheckedBy = &agentName

	a.addDomainEvent(AttendeeCheckedInEvent{
		AttendeeID: a.ID,
		RegRef:     a.RegRef,
		AgentName:  agentName,
		occurredAt: now,
	})

	return nil
}

func (a *Attendee) DomainEvents() []DomainEvent {
	return a.domainEvents
}

func (a *Attendee) ClearDomainEvents() {
	a.domainEvents = []DomainEvent{}
}

func (a *Attendee) addDomainEvent(event DomainEvent) {
	a.domainEvents = append(a.domainEvents, event)
}
