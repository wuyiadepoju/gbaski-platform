package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gbaski/gbaski-platform/app/event/domain"
	"github.com/gbaski/gbaski-ext/log"
	"github.com/gbaski/gbaski-shared/util"
	"github.com/google/uuid"
)

// CreateEventService handles creating events
type CreateEventService struct {
	eventRepo          domain.EventRepository
	categoryRepo       domain.EventCategoryRepository
	eventBus           EventBus
	webmasterService   WebmasterService
	emailScheduler     EmailScheduler
}

// NewCreateEventService creates a new CreateEventService
func NewCreateEventService(
	eventRepo domain.EventRepository,
	categoryRepo domain.EventCategoryRepository,
	eventBus EventBus,
	webmasterService WebmasterService,
	emailScheduler EmailScheduler,
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
		return nil, fmt.Errorf("invalid event name: %w", err)
	}

	eventDescription := domain.NewEventDescription(cmd.Description)

	dateRange, err := domain.NewEventDateRange(cmd.StartDate, cmd.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid date range: %w", err)
	}

	modeType, err := domain.NewEventModeType(cmd.ModeType)
	if err != nil {
		return nil, fmt.Errorf("invalid mode type: %w", err)
	}

	location := domain.NewLocation(cmd.Location, modeType)

	payment, err := domain.NewEventPayment(cmd.Payment)
	if err != nil {
		return nil, fmt.Errorf("invalid payment type: %w", err)
	}

	accessType, err := domain.NewAccessType(cmd.AccessType)
	if err != nil {
		return nil, fmt.Errorf("invalid access type: %w", err)
	}

	durationType, err := domain.NewDurationType(cmd.DurationType)
	if err != nil {
		return nil, fmt.Errorf("invalid duration type: %w", err)
	}

	registrationType, err := domain.NewRegistrationType(cmd.RegistrationType)
	if err != nil {
		return nil, fmt.Errorf("invalid registration type: %w", err)
	}

	// Handle category
	var categoryId *int
	if cmd.OtherCategory != nil && *cmd.OtherCategory != "" {
		categoryIdInt, err := s.categoryRepo.Create(domain.EventCategory{
			Name:      *cmd.OtherCategory,
			EventType: "Other",
			Public:    false,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create category: %w", err)
		}
		categoryId = &categoryIdInt
	} else if cmd.CategoryId != nil {
		categoryId = cmd.CategoryId
	}

	// Create default schema for free ticket events
	var schema json.RawMessage
	if registrationType == domain.RegistrationTypeTicket && payment == domain.EventPaymentFree {
		description := "Standard ticket at regular price"
		id := uuid.New()
		limit := 1

		tickets := []map[string]interface{}{
			{
				"id":          0,
				"uuid":        id.String(),
				"name":        "Regular",
				"description": description,
				"payment":     "free",
				"price":       0,
				"discount":    nil,
				"capacity":    nil,
				"limit":       limit,
				"sold":        0,
				"index":       0,
				"deleted":     false,
			},
		}
		schema, _ = json.Marshal(tickets)
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
		return nil, fmt.Errorf("failed to create event: %w", err)
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

// EventBus interface for publishing domain events
type EventBus interface {
	Publish(event domain.DomainEvent)
}

// WebmasterService interface for webmaster operations
type WebmasterService interface {
	AddEventURL(url string) error
}

// EmailScheduler interface for scheduling emails
type EmailScheduler interface {
	ScheduleEventEmails(event *domain.Event)
}
