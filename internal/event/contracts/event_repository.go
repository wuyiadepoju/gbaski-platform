package contracts

import (
	"encoding/json"
	"time"

	"github.com/gbaski/gbaski-platform/internal/event/domain"
	"github.com/google/uuid"
)

// QueryParams represents query parameters for event listing
type QueryParams struct {
	Search string
	Filter map[string]string
	Range  map[string][]string
	Page   int
	Size   int
}

type EventItem struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	UserID       uuid.UUID       `json:"userId" db:"user_id"`
	Name         string          `json:"name" db:"name"`
	Description  string          `json:"description" db:"description"`
	Slug         string          `json:"slug" db:"slug"`
	Status       string          `json:"status" db:"status"`
	CategoryID   *int            `json:"categoryId" db:"category_id"`
	ModeType     string          `json:"modeType" db:"mode_type"`
	Location     *string         `json:"location" db:"location"`
	Payment      string          `json:"payment" db:"payment"`
	DurationType string          `json:"durationType" db:"duration_type"`
	ImageURL     *string         `json:"imageUrl" db:"image_url"`
	VideoURL     *string         `json:"videoUrl" db:"video_url"`
	SocialMedia  json.RawMessage `json:"socialMedia" db:"social_media"`
	StartDate    time.Time       `json:"startDate" db:"start_date"`
	EndDate      time.Time       `json:"endDate" db:"end_date"`
	FormID       *uuid.UUID      `json:"formId" db:"form_id"`
	CreatedAt    time.Time       `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time       `json:"updatedAt" db:"updated_at"`
	BrandName    string          `json:"brandName" db:"brand_name"`
}

// EventList represents a paginated list of domain events returned by the repository
type EventList struct {
	Items []EventItem
	Page  int
	Size  int
	Total int
}

// EventRepository defines the contract for event persistence
type EventRepository interface {
	Save(event *domain.Event) error
	FindByID(id uuid.UUID) (*domain.Event, error)
	FindByUserID(userId uuid.UUID, query QueryParams) (EventList, error)
	Delete(id uuid.UUID) error
	UpdateImage(id uuid.UUID, imageURL string) error
	UpdateVideo(id uuid.UUID, videoURL string) error
	GetEventsReport(userId uuid.UUID) (json.RawMessage, error)
}

// EventCategoryRepository defines the contract for event category persistence
type EventCategoryRepository interface {
	Create(category domain.EventCategory) (int, error)
	FindAll() ([]domain.EventCategory, error)
}
