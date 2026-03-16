package application

import (
	"context"
	"fmt"

	"github.com/gbaski/gbaski-platform/internal/event/contracts"
	"github.com/gbaski/gbaski-platform/internal/event/domain"
	"github.com/gbaski/gbaski-shared/util"
)

// GetEventsService handles getting a list of events
type GetEventsService struct {
	eventRepo contracts.EventRepository
}

// NewGetEventsService creates a new GetEventsService
func NewGetEventsService(eventRepo contracts.EventRepository) *GetEventsService {
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

	queryParams := contracts.QueryParams{
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

	return s.toEventListDTO(eventList), nil
}

func (s *GetEventsService) toEventListDTO(eventList contracts.EventList) *EventListDTO {

	items := make([]EventItemDTO, len(eventList.Items))

	for i, item := range eventList.Items {
		items[i] = *s.eventItemToDTO(&item)
	}

	return &EventListDTO{
		Items: items,
		Page:  eventList.Page,
		Size:  eventList.Size,
		Total: eventList.Total,
	}
}

func (s *GetEventsService) eventItemToDTO(item *contracts.EventItem) *EventItemDTO {
	locationValue := item.Location
	return &EventItemDTO{
		EventDTO: EventDTO{
			ID:           item.ID,
			Name:         item.Name,
			Description:  item.Description,
			Status:       item.Status,
			StartDate:    item.StartDate,
			EndDate:      item.EndDate,
			Location:     locationValue,
			ModeType:     item.ModeType,
			Payment:      item.Payment,
			DurationType: item.DurationType,
			CategoryId:   item.CategoryID,
			ImageURL:     item.ImageURL,
			VideoURL:     item.VideoURL,
			FormId:       item.FormID,
			BrandName:    item.BrandName,
			Slug:         item.Slug,
			URL:          util.EventUrl(item.BrandName, item.Name),
			CreatedAt:    item.CreatedAt,
			UpdatedAt:    item.UpdatedAt,
		},
		// Additional fields would need to be populated from database if needed
		// For now, leaving them as zero values
	}
}

func (s *GetEventsService) eventToDTO(event *domain.Event) *EventDTO {
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
		URL:              util.EventUrl(event.BrandName(), event.Name().Value()),
		CreatedAt:        event.CreatedAt(),
		UpdatedAt:        event.UpdatedAt(),
	}
}
