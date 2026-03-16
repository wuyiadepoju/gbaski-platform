package application

import (
	"encoding/json"
	"time"

	"github.com/gbaski/gbaski-platform/internal/event/domain"
	"github.com/google/uuid"
)

// EventItem extends the Event aggregate root with additional presentation fields
type EventItem struct {
	*domain.Event
	Type              string           `json:"type"`
	Price             float64          `json:"price"`
	Currency          string           `json:"currency"`
	Category          string           `json:"category"`
	Url               string           `json:"url"`
	EngagementEnabled bool             `json:"engagementEnabled"`
	PromotionEnabled  bool             `json:"promotionEnabled"`
	Settings          *json.RawMessage `json:"settings"`
	Phone             *string          `json:"phone"`
}

// EventQueryModel represents query parameters for event listing (for HTTP parsing)
type EventQueryModel struct {
	Search string              `json:"search"`
	Filter map[string]string   `json:"filter"`
	Range  map[string][]string `json:"range"`
	Page   int                 `json:"page"`
	Size   int                 `json:"size"`
}

// CreateEventRequest represents the request to create an event
type CreateEventRequest struct {
	Name          string          `json:"name" db:"name" validate:"required"`
	Description   string          `json:"description" db:"description" validate:"required"`
	CategoryId    *int            `json:"categoryId,omitempty" db:"category_id"`
	OtherCategory *string         `json:"otherCategory,omitempty" db:"other_category"`
	AccessType    string          `json:"accessType" db:"access_type"`
	Payment       string          `json:"payment" db:"payment" validate:"required"`
	StartDate     time.Time       `json:"startDate" db:"start_date" validate:"required"`
	EndDate       time.Time       `json:"endDate" db:"end_date" validate:"required"`
	ModeType      string          `json:"modeType" db:"mode_type" validate:"required"`
	Location      *string         `json:"location,omitempty" db:"location"`
	DurationType  string          `json:"durationType" db:"duration_type" validate:"required"`
	ImageURL      string          `json:"imageUrl,omitempty" db:"image_url"`
	VideoURL      string          `json:"videoUrl,omitempty" db:"video_url"`
	RegType       string          `json:"regType" db:"reg_type" validate:"required"`
	FormId        *uuid.UUID      `json:"formId,omitempty" db:"form_id"`
	SocialMedia   json.RawMessage `json:"socialMedia,omitempty" db:"social_media"`
}

// UpdateEventImageRequest represents the request to update event image
type UpdateEventImageRequest struct {
	EventId  uuid.UUID `json:"eventId"`
	ImageURL string    `json:"imageUrl" validate:"required"`
}

// UpdateEventVideoRequest represents the request to update event video
type UpdateEventVideoRequest struct {
	EventId  uuid.UUID `json:"eventId"`
	VideoURL string    `json:"videoUrl" validate:"required"`
}

// StatusRequest represents the request to update event status
type StatusRequest struct {
	Status domain.EventStatus `json:"status" validate:"required"`
}
