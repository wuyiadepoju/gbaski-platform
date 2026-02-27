package domain

import (
	"github.com/google/uuid"
)

// EventRepository defines the interface for event persistence
type EventRepository interface {
	Save(event *Event) error
	FindByID(id uuid.UUID) (*Event, error)
	FindByUserID(userId uuid.UUID, query QueryParams) (EventList, error)
	Delete(id uuid.UUID) error
	UpdateImage(id uuid.UUID, imageURL string) error
	UpdateVideo(id uuid.UUID, videoURL string) error
}

// QueryParams represents query parameters for event listing
type QueryParams struct {
	Search string
	Filter map[string]string
	Range  map[string][]string
	Page   int
	Size   int
}

// EventList represents a paginated list of events
type EventList struct {
	Items []*Event
	Page  int
	Size  int
	Total int
}

// EventCategoryRepository defines the interface for event category persistence
type EventCategoryRepository interface {
	Create(category EventCategory) (int, error)
	FindAll() ([]EventCategory, error)
}

// EventCategory represents an event category
type EventCategory struct {
	ID        int
	Name      string
	EventType string
	Public    bool
}
