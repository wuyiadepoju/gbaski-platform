package domain

import "time"

type DomainEvent interface {
	EventType() string
	OccurredAt() time.Time
}

type SettingUpdatedEvent struct {
	Key        string
	NewValue   string
	occurredAt time.Time
}

func (e SettingUpdatedEvent) EventType() string     { return "setting.updated" }
func (e SettingUpdatedEvent) OccurredAt() time.Time { return e.occurredAt }
