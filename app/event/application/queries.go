package application

import (
	"github.com/google/uuid"
)

// GetEventQuery represents the query to get a single event
type GetEventQuery struct {
	EventID uuid.UUID
	UserID  uuid.UUID
}

// GetEventsQuery represents the query to get a list of events
type GetEventsQuery struct {
	UserID uuid.UUID
	Search string
	Filter map[string]string
	Range  map[string][]string
	Page   int
	Size   int
}

// GetEventStatsQuery represents the query to get event statistics
type GetEventStatsQuery struct {
	EventID uuid.UUID
	UserID  uuid.UUID
}

// GetEventsReportQuery represents the query to get events report
type GetEventsReportQuery struct {
	UserID uuid.UUID
}
