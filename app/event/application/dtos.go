package application

import (
	"time"
	"github.com/google/uuid"
)

// EventDTO represents the event data transfer object
type EventDTO struct {
	ID              uuid.UUID  `json:"id"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	Status          string     `json:"status"`
	StartDate       time.Time  `json:"startDate"`
	EndDate         time.Time  `json:"endDate"`
	Location        *string    `json:"location"`
	ModeType        string     `json:"modeType"`
	Payment         string     `json:"payment"`
	AccessType      string     `json:"accessType"`
	DurationType    string     `json:"durationType"`
	RegistrationType string    `json:"registrationType"`
	CategoryId      *int       `json:"categoryId"`
	ImageURL        *string    `json:"imageUrl"`
	VideoURL        *string    `json:"videoUrl"`
	FormId          *uuid.UUID `json:"formId"`
	BrandName       string     `json:"brandName"`
	Slug            string     `json:"slug"`
	URL             string     `json:"url"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// CreateEventResponseDTO represents the response after creating an event
type CreateEventResponseDTO struct {
	FormId       uuid.UUID `json:"formId"`
	EventUrl     string    `json:"eventUrl"`
	EventStatus  string    `json:"eventStatus"`
	EventPayment string    `json:"eventPayment"`
}

// EventListDTO represents a paginated list of events
type EventListDTO struct {
	Items []EventDTO `json:"items"`
	Page  int        `json:"page"`
	Size  int        `json:"size"`
	Total int        `json:"total"`
}
