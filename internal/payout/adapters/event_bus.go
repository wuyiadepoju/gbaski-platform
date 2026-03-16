package adapters

import (
	"github.com/gbaski/gbaski-ext/log"
	"github.com/gbaski/gbaski-platform/internal/payout/contracts"
	"github.com/gbaski/gbaski-platform/internal/payout/domain"
)

// EventBusImpl implements contracts.EventBus, logging domain events
type EventBusImpl struct{}

func NewEventBusImpl() contracts.EventBus {
	return &EventBusImpl{}
}

func (b *EventBusImpl) Publish(event domain.DomainEvent) {
	log.Info("payout_event_bus", "publish_event", map[string]interface{}{
		"event_type":  event.EventType(),
		"occurred_at": event.OccurredAt(),
	})
}
