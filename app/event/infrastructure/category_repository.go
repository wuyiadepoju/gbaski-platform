package infrastructure

import (
	"github.com/gbaski/gbaski-platform/app/event/domain"
	"github.com/gbaski/gbaski-event/pkg/event"
)

// EventCategoryRepositoryImpl implements the domain EventCategoryRepository interface
type EventCategoryRepositoryImpl struct {
	*event.Repository
}

// NewEventCategoryRepositoryImpl creates a new EventCategoryRepositoryImpl
func NewEventCategoryRepositoryImpl() *EventCategoryRepositoryImpl {
	return &EventCategoryRepositoryImpl{
		Repository: event.NewRepository(),
	}
}

// Create creates a new event category
func (r *EventCategoryRepositoryImpl) Create(category domain.EventCategory) (int, error) {
	var categoryId int

	err := r.DB.QueryRow(`
		INSERT INTO categories (name, event_type, public) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`,
		category.Name,
		category.EventType,
		category.Public,
	).Scan(&categoryId)

	if err != nil {
		return 0, err
	}

	return categoryId, nil
}

// FindAll finds all event categories
func (r *EventCategoryRepositoryImpl) FindAll() ([]domain.EventCategory, error) {
	categories, err := r.GetEventCategories()
	if err != nil {
		return nil, err
	}

	result := make([]domain.EventCategory, len(categories))
	for i, cat := range categories {
		result[i] = domain.EventCategory{
			ID:        cat.ID,
			Name:      cat.Name,
			EventType: cat.EventType,
			Public:    true, // Categories from GetEventCategories are public
		}
	}

	return result, nil
}
