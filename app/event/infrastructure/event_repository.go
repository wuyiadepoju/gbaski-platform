package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gbaski/gbaski-platform/app/event/domain"
	"github.com/gbaski/gbaski-event/pkg/event"
	"github.com/gbaski/gbaski-shared/pager"
	"github.com/gbaski/gbaski-shared/repo"
	"github.com/google/uuid"
)

// EventRepositoryImpl implements the domain EventRepository interface
type EventRepositoryImpl struct {
	*event.Repository
	baseRepo *repo.BaseRepository
}

// NewEventRepositoryImpl creates a new EventRepositoryImpl
func NewEventRepositoryImpl() *EventRepositoryImpl {
	return &EventRepositoryImpl{
		Repository: event.NewRepository(),
		baseRepo:    repo.NewBaseRepository(),
	}
}

// Save saves an event aggregate
func (r *EventRepositoryImpl) Save(event *domain.Event) error {
	tx, err := r.baseRepo.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if event exists
	existingEvent, _ := r.findEventByID(tx, event.ID())
	
	if existingEvent == nil {
		// Create new event
		var eventId uuid.UUID
		err = tx.QueryRow(`
			INSERT INTO events (name, description, status, slug, category_id, mode_type, location, payment, duration_type, image_url, video_url, social_media, start_date, end_date, user_id, configs)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
			RETURNING id`,
			event.Name().Value(),
			event.Description().Value(),
			event.Status().Value(),
			event.Slug(),
			event.CategoryId(),
			string(event.Location().ModeType()),
			event.Location().Value(),
			event.Payment().Value(),
			string(event.DurationType()),
			event.ImageURL(),
			event.VideoURL(),
			event.SocialMedia(),
			event.DateRange().StartDate(),
			event.DateRange().EndDate(),
			event.UserID(),
			event.Configs(),
		).Scan(&eventId)

		if err != nil {
			return err
		}

		event.SetID(eventId)

		// Create form if schema exists
		if len(event.Schema()) > 0 {
			formId, err := r.createFormForEvent(tx, event)
			if err != nil {
				return err
			}
			event.SetFormId(formId)
		}
	} else {
		// Update existing event
		_, err = tx.Exec(`
			UPDATE events 
			SET name = $1, description = $2, status = $3, slug = $4, category_id = $5, 
			    mode_type = $6, location = $7, payment = $8, duration_type = $9, 
			    image_url = $10, video_url = $11, social_media = $12, 
			    start_date = $13, end_date = $14, configs = $15, updated_at = NOW()
			WHERE id = $16 AND user_id = $17`,
			event.Name().Value(),
			event.Description().Value(),
			event.Status().Value(),
			event.Slug(),
			event.CategoryId(),
			string(event.Location().ModeType()),
			event.Location().Value(),
			event.Payment().Value(),
			string(event.DurationType()),
			event.ImageURL(),
			event.VideoURL(),
			event.SocialMedia(),
			event.DateRange().StartDate(),
			event.DateRange().EndDate(),
			event.Configs(),
			event.ID(),
			event.UserID(),
		)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// FindByID finds an event by ID
func (r *EventRepositoryImpl) FindByID(id uuid.UUID) (*domain.Event, error) {
	return r.findEventByID(nil, id)
}

func (r *EventRepositoryImpl) findEventByID(tx *sql.Tx, id uuid.UUID) (*domain.Event, error) {
	var db *sql.DB
	if tx != nil {
		db = nil // Use tx
	} else {
		db = r.baseRepo.DB
	}

	query := `
		SELECT 
			e.id, e.user_id, e.name, e.description, e.slug, e.status,
			e.category_id, e.mode_type, e.location, e.payment, e.duration_type,
			e.image_url, e.video_url, e.social_media, e.start_date, e.end_date,
			e.form_id, e.configs, e.created_at, e.updated_at,
			u.name as brand_name, f.schema
		FROM events e
		JOIN users u ON e.user_id = u.id
		LEFT JOIN forms f ON e.form_id = f.id
		WHERE e.id = $1`

	var (
		eventID          uuid.UUID
		userID           uuid.UUID
		name             string
		description      string
		slug             string
		status           string
		categoryID       *int
		modeType         string
		location         *string
		payment          string
		durationType     string
		imageURL         *string
		videoURL         *string
		socialMedia      json.RawMessage
		startDate        interface{}
		endDate          interface{}
		formID           *uuid.UUID
		configs          json.RawMessage
		createdAt        interface{}
		updatedAt        interface{}
		brandName        string
		schema           json.RawMessage
	)

	var err error
	if tx != nil {
		err = tx.QueryRow(query, id).Scan(
			&eventID, &userID, &name, &description, &slug, &status,
			&categoryID, &modeType, &location, &payment, &durationType,
			&imageURL, &videoURL, &socialMedia, &startDate, &endDate,
			&formID, &configs, &createdAt, &updatedAt, &brandName, &schema,
		)
	} else {
		err = r.baseRepo.DB.QueryRow(query, id).Scan(
			&eventID, &userID, &name, &description, &slug, &status,
			&categoryID, &modeType, &location, &payment, &durationType,
			&imageURL, &videoURL, &socialMedia, &startDate, &endDate,
			&formID, &configs, &createdAt, &updatedAt, &brandName, &schema,
		)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return r.toDomainEvent(
		eventID, userID, name, description, slug, status,
		categoryID, modeType, location, payment, durationType,
		imageURL, videoURL, socialMedia, startDate, endDate,
		formID, configs, createdAt, updatedAt, brandName,
	)
}

// FindByUserID finds events by user ID
func (r *EventRepositoryImpl) FindByUserID(userId uuid.UUID, query domain.QueryParams) (domain.EventList, error) {
	sqlString := `SELECT 
		e.id, e.name, e.description, e.slug, u.name as brand_name,
		e.category_id, e.mode_type, e.location, e.payment, e.duration_type,
		e.status, COALESCE(NULLIF(e.image_url, ''), 'https:://image.gbaski.app/gbaski/event-image.webp') AS image_url,
		e.video_url, e.start_date, e.end_date, e.form_id, e.created_at, e.updated_at
	FROM events e 
	JOIN users u ON e.user_id = u.id
	WHERE e.user_id = u.id
	ORDER BY e.created_at DESC, e.start_date DESC`

	conditions := make(map[string]interface{})

	if query.Search != "" {
		conditions["search"] = pager.SearchCondition{
			Search:  query.Search,
			Columns: []string{"e.name", "e.description"},
		}
	}

	if query.Filter != nil && len(query.Filter) > 0 {
		conditions["filter"] = pager.FilterCondition{
			Filter:  query.Filter,
			Columns: []string{"e.status"},
		}
	}

	conditions["custom"] = pager.CustomCondition{
		Items:  []string{"e.user_id = :user_id"},
		Params: map[string]any{"user_id": userId},
	}

	// Use the existing EventItem from gbaski-event package for pagination
	pager := pager.NewPager[event.EventItem]().SetQuery(sqlString).SetConditions(conditions).SetPage(query.Page)

	if query.Size > 0 {
		pager = pager.SetSize(query.Size)
	}

	result, err := pager.Query()
	if err != nil {
		return domain.EventList{}, err
	}

	// Convert to domain events
	items := make([]*domain.Event, 0)
	for _, item := range result.GetItems() {
		domainEvent, err := r.FindByID(item.Id)
		if err != nil {
			continue // Skip errors
		}
		if domainEvent != nil {
			items = append(items, domainEvent)
		}
	}

	return domain.EventList{
		Items: items,
		Page:  result.GetPage(),
		Size:  result.GetSize(),
		Total: result.GetTotal(),
	}, nil
}

// Delete deletes an event
func (r *EventRepositoryImpl) Delete(id uuid.UUID) error {
	_, err := r.baseRepo.DB.Exec(`DELETE FROM events WHERE id = $1`, id)
	return err
}

// UpdateImage updates the event image
func (r *EventRepositoryImpl) UpdateImage(id uuid.UUID, imageURL string) error {
	_, err := r.baseRepo.DB.Exec(`UPDATE events SET image_url = $1 WHERE id = $2`, imageURL, id)
	return err
}

// UpdateVideo updates the event video
func (r *EventRepositoryImpl) UpdateVideo(id uuid.UUID, videoURL string) error {
	_, err := r.baseRepo.DB.Exec(`UPDATE events SET video_url = $1 WHERE id = $2`, videoURL, id)
	return err
}

// Helper methods

func (r *EventRepositoryImpl) toDomainEvent(
	eventID uuid.UUID, userID uuid.UUID, name, description, slug, status string,
	categoryID *int, modeType string, location *string, payment, durationType string,
	imageURL, videoURL *string, socialMedia json.RawMessage,
	startDate, endDate interface{}, formID *uuid.UUID, configs, schema json.RawMessage,
	createdAt, updatedAt interface{}, brandName string,
) (*domain.Event, error) {
	// Create value objects
	eventName, err := domain.NewEventName(name)
	if err != nil {
		return nil, fmt.Errorf("invalid event name: %w", err)
	}

	eventDescription := domain.NewEventDescription(description)

	// Parse dates
	startDateParsed, ok := startDate.(time.Time)
	if !ok {
		return nil, fmt.Errorf("invalid start date type")
	}
	endDateParsed, ok := endDate.(time.Time)
	if !ok {
		return nil, fmt.Errorf("invalid end date type")
	}

	dateRange, err := domain.NewEventDateRange(startDateParsed, endDateParsed)
	if err != nil {
		return nil, fmt.Errorf("invalid date range: %w", err)
	}

	modeTypeVO, err := domain.NewEventModeType(modeType)
	if err != nil {
		return nil, fmt.Errorf("invalid mode type: %w", err)
	}

	locationValue := ""
	if location != nil {
		locationValue = *location
	}
	locationVO := domain.NewLocation(locationValue, modeTypeVO)

	paymentVO, err := domain.NewEventPayment(payment)
	if err != nil {
		return nil, fmt.Errorf("invalid payment: %w", err)
	}

	accessType, err := domain.NewAccessType("public") // Default, should be stored in DB
	if err != nil {
		return nil, fmt.Errorf("invalid access type: %w", err)
	}

	durationTypeVO, err := domain.NewDurationType(durationType)
	if err != nil {
		return nil, fmt.Errorf("invalid duration type: %w", err)
	}

	registrationType, err := domain.NewRegistrationType("ticket") // Default, should be stored in DB
	if err != nil {
		return nil, fmt.Errorf("invalid registration type: %w", err)
	}

	statusVO, err := domain.NewEventStatus(status)
	if err != nil {
		return nil, fmt.Errorf("invalid status: %w", err)
	}

	// Parse timestamps
	createdAtParsed, ok := createdAt.(time.Time)
	if !ok {
		createdAtParsed = time.Now()
	}
	updatedAtParsed, ok := updatedAt.(time.Time)
	if !ok {
		updatedAtParsed = time.Now()
	}

	// Create domain event (we'll use a factory method that doesn't raise events for existing entities)
	// For now, we'll create it manually
	event := &domain.Event{}
	// Use reflection or a factory method to set private fields
	// For simplicity, we'll need to add a factory method in the domain package

	// This reconstructs an event from persistence without raising creation events
	return r.reconstructEvent(
		eventID, userID, eventName, eventDescription, statusVO,
		dateRange, locationVO, paymentVO, accessType, durationTypeVO, registrationType,
		categoryID, imageURL, videoURL, socialMedia, configs, schema, formID, brandName, slug,
		createdAtParsed, updatedAtParsed,
	), nil
}

func (r *EventRepositoryImpl) reconstructEvent(
	id, userId uuid.UUID, name domain.EventName, description domain.EventDescription,
	status domain.EventStatus, dateRange domain.EventDateRange, location domain.Location,
	payment domain.EventPayment, accessType domain.AccessType, durationType domain.DurationType,
	registrationType domain.RegistrationType, categoryId *int, imageURL, videoURL *string,
	socialMedia, configs, schema json.RawMessage, formId *uuid.UUID, brandName, slug string,
	createdAt, updatedAt time.Time,
) *domain.Event {
	return domain.ReconstructEvent(
		id, userId, name, description, status, dateRange, location,
		payment, accessType, durationType, registrationType,
		categoryId, imageURL, videoURL, socialMedia, configs, schema,
		formId, brandName, slug, createdAt, updatedAt,
	)
}

func (r *EventRepositoryImpl) createFormForEvent(tx *sql.Tx, event *domain.Event) (uuid.UUID, error) {
	// This would need to be implemented based on the form creation logic
	// For now, return a placeholder
	return uuid.New(), nil
}
