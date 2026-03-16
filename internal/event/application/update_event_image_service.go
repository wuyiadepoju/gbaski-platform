package application

import (
	"context"
	"fmt"

	"github.com/gbaski/gbaski-platform/internal/event/contracts"
	"github.com/gbaski/gbaski-platform/internal/event/domain"
	"github.com/gbaski/gbaski-ext/log"
)

// UpdateEventImageService handles updating event images
type UpdateEventImageService struct {
	eventRepo contracts.EventRepository
	eventBus  contracts.EventBus
}

// NewUpdateEventImageService creates a new UpdateEventImageService
func NewUpdateEventImageService(
	eventRepo contracts.EventRepository,
	eventBus contracts.EventBus,
) *UpdateEventImageService {
	return &UpdateEventImageService{
		eventRepo: eventRepo,
		eventBus:  eventBus,
	}
}

// Execute executes the update event image command
func (s *UpdateEventImageService) Execute(ctx context.Context, cmd UpdateEventImageCommand) error {
	// Load existing event
	event, err := s.eventRepo.FindByID(cmd.EventID)
	if err != nil {
		return fmt.Errorf("failed to find event: %w", err)
	}

	if event == nil {
		return fmt.Errorf("event not found")
	}

	// Update image
	if err := event.UpdateImage(cmd.ImageURL); err != nil {
		return err
	}

	// Save event
	if err := s.eventRepo.Save(event); err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	// Publish domain events
	s.publishDomainEvents(event)

	log.Info("event", "update_event_image", map[string]interface{}{
		"event_id":  event.ID(),
		"image_url": cmd.ImageURL,
	})

	return nil
}

func (s *UpdateEventImageService) publishDomainEvents(event *domain.Event) {
	for _, domainEvent := range event.DomainEvents() {
		s.eventBus.Publish(domainEvent)
	}
	event.ClearDomainEvents()
}
