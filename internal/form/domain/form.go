package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AccessType string

const (
	AccessTypePublic  AccessType = "public"
	AccessTypePrivate AccessType = "private"
)

type RegistrationType string

const (
	RegistrationTypeTicket RegistrationType = "ticket"
	RegistrationTypeEntry  RegistrationType = "entry"
)

type Currency string

const (
	CurrencyNGN Currency = "NGN"
	CurrencyUSD Currency = "USD"
)

// FormAggregate represents the core form entity in the domain layer
type FormAggregate struct {
	ID            *uuid.UUID
	EventID       uuid.UUID
	Name          string
	RegType       RegistrationType
	IsPrimaryForm bool
	Price         float64
	Currency      Currency
	BuyerPaysFee  bool
	AccessType    AccessType
	Schema        any
	Configs       map[string]any
	StartDate     time.Time
	EndDate       time.Time
	UserID        uuid.UUID
}

// FormQueryModel represents the data returned for read operations
type FormQueryModel struct {
	ID        uuid.UUID        `json:"id"`
	EventID   uuid.UUID        `json:"eventId"`
	Name      string           `json:"name"`
	RegType   RegistrationType `json:"regType"`
	Price     float64          `json:"price"`
	Currency  Currency         `json:"currency"`
	Schema    json.RawMessage  `json:"schema"`
	Configs   json.RawMessage  `json:"configs"`
	EventName string           `json:"eventName"`
	FeeRate   string           `json:"feeRate"`
	CreatedAt time.Time        `json:"createdAt"`
	UpdatedAt time.Time        `json:"updatedAt"`
}
