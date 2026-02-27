package application

import (
	"context"
	"fmt"

	"github.com/gbaski/gbaski-platform/app/event/domain"
	"github.com/gbaski/gbaski-ext/log"
)

// UpdateEventStatusService handles updating event status
type UpdateEventStatusService struct {
	eventRepo domain.EventRepository
	eventBus  EventBus
}

// NewUpdateEventStatusService creates a new UpdateEventStatusService
func NewUpdateEventStatusService(
	eventRepo domain.EventRepository,
	eventBus EventBus,
) *UpdateEventStatusService {
	return &UpdateEventStatusService{
		eventRepo: eventRepo,
		eventBus:  eventBus,
	}
}

// Execute executes the update event status command
func (s *UpdateEventStatusService) Execute(ctx context.Context, cmd UpdateEventStatusCommand) (*EventDTO, error) {
	// Load existing event
	event, err := s.eventRepo.FindByID(cmd.EventID)
	if err != nil {
		return nil, fmt.Errorf("failed to find event: %w", err)
	}

	if event == nil {
		return nil, fmt.Errorf("event not found")
	}

	// Verify ownership
	if event.UserID() != cmd.UserID {
		return nil, fmt.Errorf("unauthorized access to event")
	}

	// Create status value object
	newStatus, err := domain.NewEventStatus(cmd.Status)
	if err != nil {
		return nil, fmt.Errorf("invalid status: %w", err)
	}

	// Update status
	if err := event.UpdateStatus(newStatus); err != nil {
		return nil, fmt.Errorf("failed to update status: %w", err)
	}

	// Save event
	if err := s.eventRepo.Save(event); err != nil {
		return nil, fmt.Errorf("failed to save event: %w", err)
	}

	// Publish domain events
	s.publishDomainEvents(event)

	log.Info("event", "update_event_status", map[string]interface{}{
		"event_id":  event.ID(),
		"user_id":   event.UserID(),
		"new_status": newStatus.Value(),
	})

	return s.toDTO(event), nil
}

func (s *UpdateEventStatusService) publishDomainEvents(event *domain.Event) {
	for _, domainEvent := range event.DomainEvents() {
		s.eventBus.Publish(domainEvent)
	}
	event.ClearDomainEvents()
}

func (s *UpdateEventStatusService) toDTO(event *domain.Event) *EventDTO {
	locationValue := event.Location().Value()
	return &EventDTO{
		ID:              event.ID(),
		Name:            event.Name().Value(),
		Description:     event.Description().Value(),
		Status:          event.Status().Value(),
		StartDate:       event.DateRange().StartDate(),
		EndDate:         event.DateRange().EndDate(),
		Location:        &locationValue,
		ModeType:        string(event.Location().ModeType()),
		Payment:         event.Payment().Value(),
		AccessType:      event.AccessType().Value(),
		DurationType:    string(event.DurationType()),
		RegistrationType: string(event.RegistrationType()),
		CategoryId:      event.CategoryId(),
		ImageURL:        event.ImageURL(),
		VideoURL:        event.VideoURL(),
		FormId:          event.FormId(),
		BrandName:       event.BrandName(),
		Slug:            event.Slug(),
		CreatedAt:       event.CreatedAt(),
		UpdatedAt:       event.UpdatedAt(),
	}
}
