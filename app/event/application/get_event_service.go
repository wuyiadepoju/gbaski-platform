package application

import (
	"context"
	"fmt"

	"github.com/gbaski/gbaski-platform/app/event/domain"
	"github.com/gbaski/gbaski-shared/util"
	"github.com/google/uuid"
)

// GetEventService handles getting a single event
type GetEventService struct {
	eventRepo domain.EventRepository
}

// NewGetEventService creates a new GetEventService
func NewGetEventService(eventRepo domain.EventRepository) *GetEventService {
	return &GetEventService{
		eventRepo: eventRepo,
	}
}

// Execute executes the get event query
func (s *GetEventService) Execute(ctx context.Context, query GetEventQuery) (*EventDTO, error) {
	event, err := s.eventRepo.FindByID(query.EventID)
	if err != nil {
		return nil, fmt.Errorf("failed to find event: %w", err)
	}

	if event == nil {
		return nil, fmt.Errorf("event not found")
	}

	// Verify ownership
	if event.UserID() != query.UserID {
		return nil, fmt.Errorf("unauthorized access to event")
	}

	return s.toDTO(event), nil
}

func (s *GetEventService) toDTO(event *domain.Event) *EventDTO {
	return &EventDTO{
		ID:              event.ID(),
		Name:            event.Name().Value(),
		Description:     event.Description().Value(),
		Status:          event.Status().Value(),
		StartDate:       event.DateRange().StartDate(),
		EndDate:         event.DateRange().EndDate(),
		Location:        event.Location().Value(),
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
		URL:             util.EventUrl(event.BrandName(), event.Name().Value()),
		CreatedAt:       event.CreatedAt(),
		UpdatedAt:       event.UpdatedAt(),
	}
}
