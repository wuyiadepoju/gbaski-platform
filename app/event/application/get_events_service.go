package application

import (
	"context"
	"fmt"

	"github.com/gbaski/gbaski-platform/app/event/domain"
	"github.com/gbaski/gbaski-shared/util"
	"github.com/google/uuid"
)

// GetEventsService handles getting a list of events
type GetEventsService struct {
	eventRepo domain.EventRepository
}

// NewGetEventsService creates a new GetEventsService
func NewGetEventsService(eventRepo domain.EventRepository) *GetEventsService {
	return &GetEventsService{
		eventRepo: eventRepo,
	}
}

// Execute executes the get events query
func (s *GetEventsService) Execute(ctx context.Context, query GetEventsQuery) (*EventListDTO, error) {
	// Clean up filter
	filter := query.Filter
	if filter != nil && filter["status"] == "all" {
		delete(filter, "status")
	}

	queryParams := domain.QueryParams{
		Search: query.Search,
		Filter: filter,
		Range:  query.Range,
		Page:   query.Page,
		Size:   query.Size,
	}

	eventList, err := s.eventRepo.FindByUserID(query.UserID, queryParams)
	if err != nil {
		return nil, fmt.Errorf("failed to find events: %w", err)
	}

	return s.toDTO(eventList), nil
}

func (s *GetEventsService) toDTO(eventList domain.EventList) *EventListDTO {
	items := make([]EventDTO, len(eventList.Items))
	for i, event := range eventList.Items {
		items[i] = *s.eventToDTO(event)
	}

	return &EventListDTO{
		Items: items,
		Page:  eventList.Page,
		Size:  eventList.Size,
		Total: eventList.Total,
	}
}

func (s *GetEventsService) eventToDTO(event *domain.Event) *EventDTO {
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
		URL:             util.EventUrl(event.BrandName(), event.Name().Value()),
		CreatedAt:       event.CreatedAt(),
		UpdatedAt:       event.UpdatedAt(),
	}
}
