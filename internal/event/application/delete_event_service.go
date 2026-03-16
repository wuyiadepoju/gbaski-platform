package application

import (
	"context"
	"fmt"

	"github.com/gbaski/gbaski-platform/internal/event/contracts"
	"github.com/gbaski/gbaski-platform/internal/event/domain"
	"github.com/gbaski/gbaski-ext/log"
)

// DeleteEventService handles deleting events
type DeleteEventService struct {
	eventRepo contracts.EventRepository
	eventBus  contracts.EventBus
}

// NewDeleteEventService creates a new DeleteEventService
func NewDeleteEventService(
	eventRepo contracts.EventRepository,
	eventBus contracts.EventBus,
) *DeleteEventService {
	return &DeleteEventService{
		eventRepo: eventRepo,
		eventBus:  eventBus,
	}
}

// Execute executes the delete event command
func (s *DeleteEventService) Execute(ctx context.Context, cmd DeleteEventCommand) error {
	// Load existing event
	event, err := s.eventRepo.FindByID(cmd.EventID)
	if err != nil {
		return fmt.Errorf("failed to find event: %w", err)
	}

	if event == nil {
		return fmt.Errorf("event not found")
	}

	// Verify ownership
	if event.UserID() != cmd.UserID {
		return fmt.Errorf("unauthorized access to event")
	}

	// Mark for deletion (raises domain event)
	event.Delete()

	// Delete from repository
	if err := s.eventRepo.Delete(cmd.EventID); err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	// Publish domain events
	s.publishDomainEvents(event)

	log.Info("event", "delete_event", map[string]interface{}{
		"event_id": event.ID(),
		"user_id":  event.UserID(),
	})

	return nil
}

func (s *DeleteEventService) publishDomainEvents(event *domain.Event) {
	for _, domainEvent := range event.DomainEvents() {
		s.eventBus.Publish(domainEvent)
	}
	event.ClearDomainEvents()
}
