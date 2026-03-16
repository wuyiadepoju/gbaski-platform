package contracts

import (
	"github.com/gbaski/gbaski-platform/internal/event/domain"
)

// EventBus defines the contract for publishing domain events
type EventBus interface {
	Publish(event domain.DomainEvent)
}
