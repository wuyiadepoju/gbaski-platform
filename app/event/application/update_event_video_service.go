package application

import (
	"context"
	"fmt"

	"github.com/gbaski/gbaski-platform/app/event/domain"
	"github.com/gbaski/gbaski-ext/log"
)

// UpdateEventVideoService handles updating event videos
type UpdateEventVideoService struct {
	eventRepo domain.EventRepository
	eventBus  EventBus
}

// NewUpdateEventVideoService creates a new UpdateEventVideoService
func NewUpdateEventVideoService(
	eventRepo domain.EventRepository,
	eventBus EventBus,
) *UpdateEventVideoService {
	return &UpdateEventVideoService{
		eventRepo: eventRepo,
		eventBus:  eventBus,
	}
}

// Execute executes the update event video command
func (s *UpdateEventVideoService) Execute(ctx context.Context, cmd UpdateEventVideoCommand) error {
	// Load existing event
	event, err := s.eventRepo.FindByID(cmd.EventID)
	if err != nil {
		return fmt.Errorf("failed to find event: %w", err)
	}

	if event == nil {
		return fmt.Errorf("event not found")
	}

	// Update video
	if err := event.UpdateVideo(cmd.VideoURL); err != nil {
		return fmt.Errorf("failed to update video: %w", err)
	}

	// Save event
	if err := s.eventRepo.Save(event); err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	// Publish domain events
	s.publishDomainEvents(event)

	log.Info("event", "update_event_video", map[string]interface{}{
		"event_id":  event.ID(),
		"video_url": cmd.VideoURL,
	})

	return nil
}

func (s *UpdateEventVideoService) publishDomainEvents(event *domain.Event) {
	for _, domainEvent := range event.DomainEvents() {
		s.eventBus.Publish(domainEvent)
	}
	event.ClearDomainEvents()
}
