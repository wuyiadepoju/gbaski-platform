package repo

import (
	"github.com/gbaski/gbaski-platform/internal/event/contracts"
	"github.com/gbaski/gbaski-platform/internal/event/domain"
	"github.com/gbaski/gbaski-shared/repo"
	"github.com/jmoiron/sqlx"
)

// EventCategoryRepositoryImpl implements the contracts.EventCategoryRepository interface
type EventCategoryRepositoryImpl struct {
	baseRepo *repo.BaseRepository
}

// NewEventCategoryRepositoryImpl creates a new EventCategoryRepositoryImpl
// If db is nil, uses BaseRepository's default DB connection
func NewEventCategoryRepositoryImpl(db *sqlx.DB) contracts.EventCategoryRepository {
	baseRepo := repo.NewBaseRepository()
	if db != nil {
		baseRepo.DB = db
	}
	return &EventCategoryRepositoryImpl{
		baseRepo: baseRepo,
	}
}

// Create creates a new event category
func (r *EventCategoryRepositoryImpl) Create(category domain.EventCategory) (int, error) {
	var categoryId int

	err := r.baseRepo.DB.QueryRow(`
		INSERT INTO categories (name, event_type, public) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`,
		category.GetName(),
		category.GetEventType(),
		category.IsPublic(),
	).Scan(&categoryId)

	if err != nil {
		return 0, err
	}

	return categoryId, nil
}

// FindAll finds all event categories
func (r *EventCategoryRepositoryImpl) FindAll() ([]domain.EventCategory, error) {
	rows, err := r.baseRepo.DB.Query(`
		SELECT id, name, event_type, public 
		FROM categories 
		WHERE public = true 
		ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []domain.EventCategory
	for rows.Next() {
		var id int
		var name, eventType string
		var public bool
		if err := rows.Scan(&id, &name, &eventType, &public); err != nil {
			return nil, err
		}
		// Reconstruct category from persistence
		cat, err := domain.NewEventCategory(name, eventType, public)
		if err != nil {
			return nil, err
		}
		cat = cat.WithID(id)
		categories = append(categories, cat)
	}

	return categories, rows.Err()
}
