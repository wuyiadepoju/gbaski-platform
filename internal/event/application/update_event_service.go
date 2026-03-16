package application

import (
	"context"
	"fmt"

	"github.com/gbaski/gbaski-ext/log"
	"github.com/gbaski/gbaski-platform/internal/event/contracts"
	"github.com/gbaski/gbaski-platform/internal/event/domain"
)

// UpdateEventService handles updating events
type UpdateEventService struct {
	eventRepo      contracts.EventRepository
	categoryRepo   contracts.EventCategoryRepository
	eventBus       contracts.EventBus
	emailScheduler contracts.EmailScheduler
}

// NewUpdateEventService creates a new UpdateEventService
func NewUpdateEventService(
	eventRepo contracts.EventRepository,
	categoryRepo contracts.EventCategoryRepository,
	eventBus contracts.EventBus,
	emailScheduler contracts.EmailScheduler,
) *UpdateEventService {
	return &UpdateEventService{
		eventRepo:      eventRepo,
		categoryRepo:   categoryRepo,
		eventBus:       eventBus,
		emailScheduler: emailScheduler,
	}
}

// Execute executes the update event command
func (s *UpdateEventService) Execute(ctx context.Context, cmd UpdateEventCommand) (*EventDTO, error) {
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

	// Create value objects
	eventName, err := domain.NewEventName(cmd.Name)
	if err != nil {
		return nil, err
	}

	eventDescription := domain.NewEventDescription(cmd.Description)

	dateRange, err := domain.NewEventDateRange(cmd.StartDate, cmd.EndDate)
	if err != nil {
		return nil, err
	}

	// Check if dates changed
	oldDateRange := event.DateRange()
	datesChanged := !oldDateRange.StartDate().Equal(dateRange.StartDate()) || !oldDateRange.EndDate().Equal(dateRange.EndDate())

	// Update event
	if err := event.UpdateName(eventName); err != nil {
		return nil, err
	}

	if err := event.UpdateDescription(eventDescription); err != nil {
		return nil, err
	}

	if datesChanged {
		if err := event.UpdateDates(dateRange); err != nil {
			return nil, err
		}
		// Reschedule emails if dates changed
		go s.emailScheduler.RescheduleEventEmails(event, dateRange.StartDate(), dateRange.EndDate())
	}

	// Handle category
	if cmd.OtherCategory != nil && *cmd.OtherCategory != "" {
		// Create new category using domain factory
		category, err := domain.NewEventCategory(*cmd.OtherCategory, "Other", false)
		if err != nil {
			return nil, fmt.Errorf("failed to create category: %w", err)
		}
		categoryIdInt, err := s.categoryRepo.Create(category)
		if err != nil {
			return nil, fmt.Errorf("failed to save category: %w", err)
		}
		event.SetCategoryId(categoryIdInt)
	} else if cmd.CategoryId != nil {
		event.SetCategoryId(*cmd.CategoryId)
	}

	// Note: Location update would require adding an UpdateLocation method to the Event aggregate
	// For now, location updates are handled through the full event update

	event.MarkAsUpdated()

	// Save event
	if err := s.eventRepo.Save(event); err != nil {
		return nil, fmt.Errorf("failed to save event: %w", err)
	}

	// Publish domain events
	s.publishDomainEvents(event)

	log.Info("event", "update_event", map[string]interface{}{
		"event_id": event.ID(),
		"user_id":  event.UserID(),
	})

	return s.toDTO(event), nil
}

func (s *UpdateEventService) publishDomainEvents(event *domain.Event) {
	for _, domainEvent := range event.DomainEvents() {
		s.eventBus.Publish(domainEvent)
	}
	event.ClearDomainEvents()
}

func (s *UpdateEventService) toDTO(event *domain.Event) *EventDTO {
	locationValue := event.Location().Value()
	return &EventDTO{
		ID:               event.ID(),
		Name:             event.Name().Value(),
		Description:      event.Description().Value(),
		Status:           event.Status().Value(),
		StartDate:        event.DateRange().StartDate(),
		EndDate:          event.DateRange().EndDate(),
		Location:         &locationValue,
		ModeType:         string(event.Location().ModeType()),
		Payment:          event.Payment().Value(),
		AccessType:       event.AccessType().Value(),
		DurationType:     string(event.DurationType()),
		RegistrationType: string(event.RegistrationType()),
		CategoryId:       event.CategoryId(),
		ImageURL:         event.ImageURL(),
		VideoURL:         event.VideoURL(),
		FormId:           event.FormId(),
		BrandName:        event.BrandName(),
		Slug:             event.Slug(),
		CreatedAt:        event.CreatedAt(),
		UpdatedAt:        event.UpdatedAt(),
	}
}
