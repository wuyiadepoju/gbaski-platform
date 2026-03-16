package application

import (
	"context"

	"github.com/gbaski/gbaski-ext/log"
	"github.com/gbaski/gbaski-platform/internal/form/contracts"
	"github.com/google/uuid"
)

// GetFormsQuery represents the query to fetch forms by event
type GetFormsQuery struct {
	EventID uuid.UUID
}

// GetFormsService handles retrieval of forms
type GetFormsService struct {
	repo contracts.FormRepository
}

func NewGetFormsService(repo contracts.FormRepository) *GetFormsService {
	return &GetFormsService{
		repo: repo,
	}
}

func (s *GetFormsService) Execute(ctx context.Context, query GetFormsQuery) (interface{}, error) {
	forms, err := s.repo.GetFormsByEventId(ctx, query.EventID)
	if err != nil {
		log.Error(formComponent, "get_forms_by_event_id", err)
		return nil, err
	}

	log.Info(formComponent, "get_forms_by_event_id", map[string]interface{}{
		"event_id":   query.EventID,
		"forms_count": len(forms),
	})

	return forms, nil
}

