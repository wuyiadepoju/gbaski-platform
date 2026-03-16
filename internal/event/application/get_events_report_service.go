package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gbaski/gbaski-platform/internal/event/contracts"
)

// GetEventsReportService handles getting events report
type GetEventsReportService struct {
	eventRepo contracts.EventRepository
}

// NewGetEventsReportService creates a new GetEventsReportService
func NewGetEventsReportService(eventRepo contracts.EventRepository) *GetEventsReportService {
	return &GetEventsReportService{
		eventRepo: eventRepo,
	}
}

// Execute executes the get events report query
func (s *GetEventsReportService) Execute(ctx context.Context, query GetEventsReportQuery) (*EventsReportDTO, error) {
	reportData, err := s.eventRepo.GetEventsReport(query.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get events report: %w", err)
	}

	// Parse the JSON response from the database function
	var report EventsReportDTO
	if err := json.Unmarshal(reportData, &report); err != nil {
		return nil, fmt.Errorf("failed to parse events report: %w", err)
	}

	return &report, nil
}
