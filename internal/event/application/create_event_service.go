package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gbaski/gbaski-ext/log"
	"github.com/gbaski/gbaski-platform/internal/event/contracts"
	"github.com/gbaski/gbaski-platform/internal/event/domain"
	"github.com/gbaski/gbaski-shared/util"
)

// CreateEventService handles creating events
type CreateEventService struct {
	eventRepo        contracts.EventRepository
	categoryRepo     contracts.EventCategoryRepository
	eventBus         contracts.EventBus
	webmasterService contracts.WebmasterService
	emailScheduler   contracts.EmailScheduler
}

// NewCreateEventService creates a new CreateEventService
func NewCreateEventService(
	eventRepo contracts.EventRepository,
	categoryRepo contracts.EventCategoryRepository,
	eventBus contracts.EventBus,
	webmasterService contracts.WebmasterService,
	emailScheduler contracts.EmailScheduler,
) *CreateEventService {
	return &CreateEventService{
		eventRepo:        eventRepo,
		categoryRepo:     categoryRepo,
		eventBus:         eventBus,
		webmasterService: webmasterService,
		emailScheduler:   emailScheduler,
	}
}

// Execute executes the create event command
func (s *CreateEventService) Execute(ctx context.Context, cmd CreateEventCommand) (*CreateEventResponseDTO, error) {
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

	modeType, err := domain.NewEventModeType(cmd.ModeType)
	if err != nil {
		return nil, err
	}

	locationValue := ""
	if cmd.Location != nil {
		locationValue = *cmd.Location
	}
	// Use NewLocationWithDefaults to handle empty locations with business defaults
	location := domain.NewLocationWithDefaults(locationValue, modeType)

	payment, err := domain.NewEventPayment(cmd.Payment)
	if err != nil {
		return nil, err
	}

	accessType, err := domain.NewAccessType(cmd.AccessType)
	if err != nil {
		return nil, err
	}

	durationType, err := domain.NewDurationType(cmd.DurationType)
	if err != nil {
		return nil, err
	}

	registrationType, err := domain.NewRegistrationType(cmd.RegistrationType)
	if err != nil {
		return nil, err
	}

	// Handle category
	var categoryId *int
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
		categoryId = &categoryIdInt
	} else if cmd.CategoryId != nil {
		categoryId = cmd.CategoryId
	}

	// Create default schema for free ticket events using domain service
	var schema json.RawMessage
	schemaService := domain.NewEventSchemaService()
	if schemaService.ShouldCreateDefaultSchema(registrationType, payment) {
		schema, err = schemaService.CreateDefaultFreeTicketSchema()
		if err != nil {
			return nil, fmt.Errorf("failed to create default schema: %w", err)
		}
	}

	// Create domain entity
	event, err := domain.NewEvent(
		cmd.UserID,
		eventName,
		eventDescription,
		dateRange,
		location,
		payment,
		accessType,
		durationType,
		registrationType,
		cmd.BrandName,
	)
	if err != nil {
		return nil, err
	}

	if categoryId != nil {
		event.SetCategoryId(*categoryId)
	}

	if len(schema) > 0 {
		event.SetSchema(schema)
	}

	// Save event
	if err := s.eventRepo.Save(event); err != nil {
		return nil, fmt.Errorf("failed to save event: %w", err)
	}

	// Publish domain events
	s.publishDomainEvents(event)

	// Schedule webmaster submission (async)
	go s.submitToWebmaster(event, cmd.BrandName)

	// Schedule email notifications (async)
	go s.emailScheduler.ScheduleEventEmails(event)

	log.Info("event", "create_event", map[string]interface{}{
		"event_id": event.ID(),
		"user_id":  event.UserID(),
		"form_id":  event.FormId(),
	})

	// Return response
	formId := event.FormId()
	if formId == nil {
		return nil, fmt.Errorf("form ID is nil after event creation")
	}

	return &CreateEventResponseDTO{
		FormId:       *formId,
		EventUrl:     util.EventUrl(cmd.BrandName, event.Name().Value()),
		EventStatus:  event.Status().Value(),
		EventPayment: event.Payment().Value(),
	}, nil
}

func (s *CreateEventService) publishDomainEvents(event *domain.Event) {
	for _, domainEvent := range event.DomainEvents() {
		s.eventBus.Publish(domainEvent)
	}
	event.ClearDomainEvents()
}

func (s *CreateEventService) submitToWebmaster(event *domain.Event, brandName string) {
	if err := s.webmasterService.AddEventURL(util.EventUrl(brandName, event.Name().Value())); err != nil {
		log.Error("webmaster", "add_event_url", err)
	}
}
