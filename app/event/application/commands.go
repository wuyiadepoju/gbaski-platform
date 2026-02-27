package application

import (
	"time"
	"github.com/google/uuid"
)

// CreateEventCommand represents the command to create an event
type CreateEventCommand struct {
	UserID          uuid.UUID
	Name            string
	Description     string
	StartDate       time.Time
	EndDate         time.Time
	Location        *string
	ModeType        string
	Payment         string
	AccessType      string
	DurationType    string
	RegistrationType string
	CategoryId      *int
	OtherCategory   *string
	BrandName        string
}

// UpdateEventCommand represents the command to update an event
type UpdateEventCommand struct {
	EventID         uuid.UUID
	UserID          uuid.UUID
	Name            string
	Description     string
	StartDate       time.Time
	EndDate         time.Time
	Location        *string
	ModeType        string
	DurationType    string
	CategoryId      *int
	OtherCategory   *string
}

// UpdateEventStatusCommand represents the command to update event status
type UpdateEventStatusCommand struct {
	EventID uuid.UUID
	UserID  uuid.UUID
	Status  string
}

// UpdateEventImageCommand represents the command to update event image
type UpdateEventImageCommand struct {
	EventID  uuid.UUID
	ImageURL string
}

// UpdateEventVideoCommand represents the command to update event video
type UpdateEventVideoCommand struct {
	EventID  uuid.UUID
	VideoURL string
}

// DeleteEventCommand represents the command to delete an event
type DeleteEventCommand struct {
	EventID uuid.UUID
	UserID  uuid.UUID
}
