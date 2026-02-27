package form

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

// FormRequest represents the data required to create or update a form
type FormRequest struct {
	Id            *uuid.UUID       `json:"id" db:"id"`
	EventId       uuid.UUID        `json:"eventId" db:"event_id"`
	Name          string           `json:"name" db:"name"`
	RegType       RegistrationType `json:"regType" db:"reg_type"`
	IsPrimaryForm bool             `json:"isPrimaryForm" db:"is_primary_form"`
	Price         float64          `json:"price" db:"price"`
	Currency      Currency         `json:"currency" db:"currency"`
	BuyerPaysFee  bool             `json:"buyerPaysFee" db:"buyer_pays_fee"`
	AccessType    AccessType       `json:"accessType" db:"access_type"`
	Schema        any              `json:"schema" db:"schema"`
	Configs       map[string]any   `json:"configs" db:"configs"`
	StartDate     time.Time        `json:"startDate" db:"start_date"`
	EndDate       time.Time        `json:"endDate" db:"end_date"`
}

type FormModel struct {
	FormRequest
	Price   float64         `json:"price" db:"price"`
	Schema  json.RawMessage `json:"schema" db:"schema"`
	Configs json.RawMessage `json:"configs" db:"configs"`
	UserId  uuid.UUID       `json:"userId" db:"user_id"`
}

type FormItem struct {
	FormModel
	Id        uuid.UUID `json:"id" db:"id"`
	EventName string    `json:"eventName" db:"event_name"`
	FeeRate   string    `json:"feeRate" db:"fee_rate"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}
