package domain

import (
	"strings"
)

// EventCategory represents an event category entity
// This is a reference entity in the Event bounded context
// Categories are shared across events and have their own lifecycle
type EventCategory struct {
	Id        int
	Name      string
	EventType string
	Public    bool
}

// NewEventCategory creates a new EventCategory entity
func NewEventCategory(name, eventType string, public bool) (EventCategory, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return EventCategory{}, ErrEmptyCategoryName
	}
	if len(name) > 200 {
		return EventCategory{}, ErrCategoryNameTooLong
	}

	eventType = strings.TrimSpace(eventType)
	if eventType == "" {
		return EventCategory{}, ErrEmptyEventType
	}

	return EventCategory{
		Name:      name,
		EventType: eventType,
		Public:    public,
	}, nil
}

// WithID sets the ID for an existing category (used by repository)
func (c EventCategory) WithID(id int) EventCategory {
	c.Id = id
	return c
}

// ID returns the category ID
func (c EventCategory) ID() int {
	return c.Id
}

// GetName returns the category name
func (c EventCategory) GetName() string {
	return c.Name
}

// GetEventType returns the event type
func (c EventCategory) GetEventType() string {
	return c.EventType
}

// IsPublic returns whether the category is public
func (c EventCategory) IsPublic() bool {
	return c.Public
}
