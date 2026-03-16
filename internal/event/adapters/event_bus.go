package adapters

import (
	"github.com/gbaski/gbaski-platform/internal/event/contracts"
	"github.com/gbaski/gbaski-platform/internal/event/domain"
	"github.com/gbaski/gbaski-ext/log"
)

// EventBusImpl implements the contracts.EventBus interface
// In a real implementation, this would publish to a message queue (RabbitMQ, Kafka, etc.)
type EventBusImpl struct {
	// In a real implementation, this would have a message queue client
}

// NewEventBusImpl creates a new EventBusImpl
func NewEventBusImpl() contracts.EventBus {
	return &EventBusImpl{}
}

// Publish publishes a domain event
func (b *EventBusImpl) Publish(event domain.DomainEvent) {
	// In a real implementation, this would publish to a message queue
	// For now, we just log it
	log.Info("event_bus", "publish_event", map[string]interface{}{
		"event_type":  event.EventType(),
		"occurred_at": event.OccurredAt(),
	})
}
