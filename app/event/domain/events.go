package domain

import (
	"time"
	"github.com/google/uuid"
)

// DomainEvent represents a domain event
type DomainEvent interface {
	EventType() string
	OccurredAt() time.Time
}

// EventCreatedEvent is raised when an event is created
type EventCreatedEvent struct {
	EventID    uuid.UUID
	UserID     uuid.UUID
	EventName  string
	OccurredAt time.Time
}

func (e EventCreatedEvent) EventType() string {
	return "event.created"
}

// EventPublishedEvent is raised when an event is published
type EventPublishedEvent struct {
	EventID    uuid.UUID
	UserID     uuid.UUID
	OccurredAt time.Time
}

func (e EventPublishedEvent) EventType() string {
	return "event.published"
}

// EventUpdatedEvent is raised when an event is updated
type EventUpdatedEvent struct {
	EventID    uuid.UUID
	UserID     uuid.UUID
	OccurredAt time.Time
}

func (e EventUpdatedEvent) EventType() string {
	return "event.updated"
}

// EventStatusChangedEvent is raised when event status changes
type EventStatusChangedEvent struct {
	EventID    uuid.UUID
	UserID     uuid.UUID
	OldStatus  EventStatus
	NewStatus  EventStatus
	OccurredAt time.Time
}

func (e EventStatusChangedEvent) EventType() string {
	return "event.status.changed"
}

// EventDeletedEvent is raised when an event is deleted
type EventDeletedEvent struct {
	EventID    uuid.UUID
	UserID     uuid.UUID
	OccurredAt time.Time
}

func (e EventDeletedEvent) EventType() string {
	return "event.deleted"
}

// EventDatesChangedEvent is raised when event dates are changed
type EventDatesChangedEvent struct {
	EventID      uuid.UUID
	UserID       uuid.UUID
	OldStartDate time.Time
	NewStartDate time.Time
	OldEndDate   time.Time
	NewEndDate   time.Time
	OccurredAt   time.Time
}

func (e EventDatesChangedEvent) EventType() string {
	return "event.dates.changed"
}
