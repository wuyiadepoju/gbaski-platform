package application

import (
	"time"

	"github.com/gbaski/gbaski-platform/internal/form/domain"
	"github.com/google/uuid"
)

// CreateFormRequest represents the incoming HTTP request payload for creating/updating a form
type CreateFormRequest struct {
	Id            *uuid.UUID              `json:"id"`
	EventId       uuid.UUID               `json:"eventId" validate:"required"`
	Name          string                  `json:"name" validate:"required"`
	RegType       domain.RegistrationType `json:"regType" validate:"required"`
	IsPrimaryForm bool                    `json:"isPrimaryForm"`
	Price         float64                 `json:"price"`
	Currency      domain.Currency         `json:"currency"`
	BuyerPaysFee  bool                    `json:"buyerPaysFee"`
	AccessType    domain.AccessType       `json:"accessType"`
	Schema        any                     `json:"schema"`
	Configs       map[string]any          `json:"configs"`
	StartDate     time.Time               `json:"startDate"`
	EndDate       time.Time               `json:"endDate"`
}

