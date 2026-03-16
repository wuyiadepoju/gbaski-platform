package domain

import (
	"time"
)

type Setting struct {
	Key       string
	Value     string
	Category  string
	CreatedAt time.Time
	UpdatedAt time.Time

	// Domain events
	domainEvents []DomainEvent
}

func NewSetting(key, value, category string) (*Setting, error) {
	if key == "" {
		return nil, ErrInvalidKey
	}

	return &Setting{
		Key:          key,
		Value:        value,
		Category:     category,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		domainEvents: []DomainEvent{},
	}, nil
}

func (s *Setting) UpdateValue(newValue string) {
	s.Value = newValue
	s.UpdatedAt = time.Now()
	s.addDomainEvent(SettingUpdatedEvent{
		Key:        s.Key,
		NewValue:   newValue,
		occurredAt: s.UpdatedAt,
	})
}

func (s *Setting) DomainEvents() []DomainEvent {
	return s.domainEvents
}

func (s *Setting) ClearDomainEvents() {
	s.domainEvents = []DomainEvent{}
}

func (s *Setting) addDomainEvent(event DomainEvent) {
	s.domainEvents = append(s.domainEvents, event)
}
