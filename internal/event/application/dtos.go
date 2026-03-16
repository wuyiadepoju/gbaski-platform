package application

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EventDTO represents the event data transfer object
type EventDTO struct {
	ID               uuid.UUID  `json:"id"`
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	Status           string     `json:"status"`
	StartDate        time.Time  `json:"startDate"`
	EndDate          time.Time  `json:"endDate"`
	Location         *string    `json:"location"`
	ModeType         string     `json:"modeType"`
	Payment          string     `json:"payment"`
	AccessType       string     `json:"accessType"`
	DurationType     string     `json:"durationType"`
	RegistrationType string     `json:"registrationType"`
	CategoryId       *int       `json:"categoryId"`
	ImageURL         *string    `json:"imageUrl"`
	VideoURL         *string    `json:"videoUrl"`
	FormId           *uuid.UUID `json:"formId"`
	BrandName        string     `json:"brandName"`
	Slug             string     `json:"slug"`
	URL              string     `json:"url"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

type EventItemDTO struct {
	EventDTO
	Type              string           `json:"type"`
	Price             float64          `json:"price"`
	Currency          string           `json:"currency"`
	Category          string           `json:"category"`
	Url               string           `json:"url"`
	EngagementEnabled bool             `json:"engagementEnabled"`
	PromotionEnabled  bool             `json:"promotionEnabled"`
	Settings          *json.RawMessage `json:"settings"`
	Phone             *string          `json:"phone"`
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
	Items []EventItemDTO `json:"items"`
	Page  int            `json:"page"`
	Size  int            `json:"size"`
	Total int            `json:"total"`
}

// EventDetailDTO represents detailed event information with features
type EventDetailDTO struct {
	Item     EventItemDTO      `json:"item"`
	Features []EventFeatureDTO `json:"features"`
}

// EventFeatureDTO represents an event feature
type EventFeatureDTO struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Tag         string    `json:"tag" db:"tag"`
	AccessType  string    `json:"accessType" db:"access_type"`
	FormId      uuid.UUID `json:"formId" db:"form_id"`
}

// EventStatsItemDTO represents event statistics item
type EventStatsItemDTO struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Currency    string  `json:"currency"`
	Quantity    int     `json:"quantity"`
	Amount      float64 `json:"amount"`
}

// EventSummaryDTO represents a summary of an event
type EventSummaryDTO struct {
	EventID   string `json:"event_id"`
	EventName string `json:"event_name"`
	EndDate   string `json:"end_date"`
}

// UpcomingEventSummaryDTO represents upcoming events summary
type UpcomingEventSummaryDTO struct {
	UpcomingEventsCount int    `json:"upcoming_events_count"`
	NextEventDate       string `json:"next_event_date"`
	NextEventID         string `json:"next_event_id"`
}

// EventsReportDTO represents events report data
type EventsReportDTO struct {
	CurrentActiveEvent *EventSummaryDTO         `json:"current_active_event"`
	MostRecentEvent    *EventSummaryDTO         `json:"most_recent_event"`
	UpcomingEvents     *UpcomingEventSummaryDTO `json:"upcoming_events"`
}
