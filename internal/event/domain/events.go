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
	eventID    uuid.UUID
	userID     uuid.UUID
	eventName  string
	occurredAt time.Time
}

func NewEventCreatedEvent(eventID, userID uuid.UUID, eventName string) EventCreatedEvent {
	return EventCreatedEvent{
		eventID:    eventID,
		userID:     userID,
		eventName:  eventName,
		occurredAt: time.Now(),
	}
}

func (e EventCreatedEvent) EventType() string {
	return "event.created"
}

func (e EventCreatedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e EventCreatedEvent) EventID() uuid.UUID {
	return e.eventID
}

func (e EventCreatedEvent) UserID() uuid.UUID {
	return e.userID
}

func (e EventCreatedEvent) EventName() string {
	return e.eventName
}

// EventPublishedEvent is raised when an event is published
type EventPublishedEvent struct {
	eventID    uuid.UUID
	userID     uuid.UUID
	occurredAt time.Time
}

func NewEventPublishedEvent(eventID, userID uuid.UUID) EventPublishedEvent {
	return EventPublishedEvent{
		eventID:    eventID,
		userID:     userID,
		occurredAt: time.Now(),
	}
}

func (e EventPublishedEvent) EventType() string {
	return "event.published"
}

func (e EventPublishedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e EventPublishedEvent) EventID() uuid.UUID {
	return e.eventID
}

func (e EventPublishedEvent) UserID() uuid.UUID {
	return e.userID
}

// EventUpdatedEvent is raised when an event is updated
type EventUpdatedEvent struct {
	eventID    uuid.UUID
	userID     uuid.UUID
	occurredAt time.Time
}

func NewEventUpdatedEvent(eventID, userID uuid.UUID) EventUpdatedEvent {
	return EventUpdatedEvent{
		eventID:    eventID,
		userID:     userID,
		occurredAt: time.Now(),
	}
}

func (e EventUpdatedEvent) EventType() string {
	return "event.updated"
}

func (e EventUpdatedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e EventUpdatedEvent) EventID() uuid.UUID {
	return e.eventID
}

func (e EventUpdatedEvent) UserID() uuid.UUID {
	return e.userID
}

// EventStatusChangedEvent is raised when event status changes
type EventStatusChangedEvent struct {
	eventID    uuid.UUID
	userID     uuid.UUID
	oldStatus  EventStatus
	newStatus  EventStatus
	occurredAt time.Time
}

func NewEventStatusChangedEvent(eventID, userID uuid.UUID, oldStatus, newStatus EventStatus) EventStatusChangedEvent {
	return EventStatusChangedEvent{
		eventID:    eventID,
		userID:     userID,
		oldStatus:  oldStatus,
		newStatus:  newStatus,
		occurredAt: time.Now(),
	}
}

func (e EventStatusChangedEvent) EventType() string {
	return "event.status.changed"
}

func (e EventStatusChangedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e EventStatusChangedEvent) EventID() uuid.UUID {
	return e.eventID
}

func (e EventStatusChangedEvent) UserID() uuid.UUID {
	return e.userID
}

func (e EventStatusChangedEvent) OldStatus() EventStatus {
	return e.oldStatus
}

func (e EventStatusChangedEvent) NewStatus() EventStatus {
	return e.newStatus
}

// EventDeletedEvent is raised when an event is deleted
type EventDeletedEvent struct {
	eventID    uuid.UUID
	userID     uuid.UUID
	occurredAt time.Time
}

func NewEventDeletedEvent(eventID, userID uuid.UUID) EventDeletedEvent {
	return EventDeletedEvent{
		eventID:    eventID,
		userID:     userID,
		occurredAt: time.Now(),
	}
}

func (e EventDeletedEvent) EventType() string {
	return "event.deleted"
}

func (e EventDeletedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e EventDeletedEvent) EventID() uuid.UUID {
	return e.eventID
}

func (e EventDeletedEvent) UserID() uuid.UUID {
	return e.userID
}

// EventDatesChangedEvent is raised when event dates are changed
type EventDatesChangedEvent struct {
	eventID      uuid.UUID
	userID       uuid.UUID
	oldStartDate time.Time
	newStartDate time.Time
	oldEndDate   time.Time
	newEndDate   time.Time
	occurredAt   time.Time
}

func NewEventDatesChangedEvent(eventID, userID uuid.UUID, oldStartDate, newStartDate, oldEndDate, newEndDate time.Time) EventDatesChangedEvent {
	return EventDatesChangedEvent{
		eventID:      eventID,
		userID:       userID,
		oldStartDate: oldStartDate,
		newStartDate: newStartDate,
		oldEndDate:   oldEndDate,
		newEndDate:   newEndDate,
		occurredAt:   time.Now(),
	}
}

func (e EventDatesChangedEvent) EventType() string {
	return "event.dates.changed"
}

func (e EventDatesChangedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e EventDatesChangedEvent) EventID() uuid.UUID {
	return e.eventID
}

func (e EventDatesChangedEvent) UserID() uuid.UUID {
	return e.userID
}

func (e EventDatesChangedEvent) OldStartDate() time.Time {
	return e.oldStartDate
}

func (e EventDatesChangedEvent) NewStartDate() time.Time {
	return e.newStartDate
}

func (e EventDatesChangedEvent) OldEndDate() time.Time {
	return e.oldEndDate
}

func (e EventDatesChangedEvent) NewEndDate() time.Time {
	return e.newEndDate
}
