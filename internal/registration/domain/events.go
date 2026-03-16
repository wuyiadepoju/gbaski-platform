package domain

import (
	"time"

	"github.com/google/uuid"
)

type DomainEvent interface {
	EventType() string
	OccurredAt() time.Time
}

type AttendeeCheckedInEvent struct {
	AttendeeID int64
	RegRef     string
	AgentName  string
	occurredAt time.Time
}

func (e AttendeeCheckedInEvent) EventType() string     { return "registration.attendee.checked_in" }
func (e AttendeeCheckedInEvent) OccurredAt() time.Time { return e.occurredAt }

type RegistrationCreatedEvent struct {
	FormID     uuid.UUID
	RegRef     string
	occurredAt time.Time
}

func (e RegistrationCreatedEvent) EventType() string     { return "registration.created" }
func (e RegistrationCreatedEvent) OccurredAt() time.Time { return e.occurredAt }
