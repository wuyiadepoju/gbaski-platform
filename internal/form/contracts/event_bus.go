package contracts

import (
	eventdomain "github.com/gbaski/gbaski-platform/internal/event/domain"
)

// EventBus defines the contract for publishing domain events related to forms.
// It reuses the generic DomainEvent interface from the event domain package.
type EventBus interface {
	Publish(event eventdomain.DomainEvent)
}

