package domain

import (
	"encoding/json"

	"github.com/google/uuid"
)

// EventSchemaService is a domain service for creating event schemas
// This encapsulates business logic for schema creation that doesn't naturally fit in the Event aggregate
type EventSchemaService struct{}

// NewEventSchemaService creates a new EventSchemaService
func NewEventSchemaService() *EventSchemaService {
	return &EventSchemaService{}
}

// CreateDefaultFreeTicketSchema creates a default schema for free ticket events
// This is business logic that determines what schema should be created for free ticket events
func (s *EventSchemaService) CreateDefaultFreeTicketSchema() (json.RawMessage, error) {
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

	return json.Marshal(tickets)
}

// ShouldCreateDefaultSchema determines if a default schema should be created
// This encapsulates the business rule: free ticket events get a default schema
func (s *EventSchemaService) ShouldCreateDefaultSchema(registrationType RegistrationType, payment EventPayment) bool {
	return registrationType == RegistrationTypeTicket && payment == EventPaymentFree
}
