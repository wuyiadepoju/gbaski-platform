package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gbaski/gbaski-platform/internal/event/contracts"
	"github.com/gbaski/gbaski-platform/internal/event/domain"
	"github.com/gbaski/gbaski-shared/pager"
	"github.com/gbaski/gbaski-shared/repo"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// eventRowDTO represents a database row for an event
type eventRowDTO struct {
	EventID      uuid.UUID       `db:"id"`
	UserID       uuid.UUID       `db:"user_id"`
	Name         string          `db:"name"`
	Description  string          `db:"description"`
	Slug         string          `db:"slug"`
	Status       string          `db:"status"`
	CategoryID   *int            `db:"category_id"`
	ModeType     string          `db:"mode_type"`
	Location     *string         `db:"location"`
	Payment      string          `db:"payment"`
	DurationType string          `db:"duration_type"`
	ImageURL     *string         `db:"image_url"`
	VideoURL     *string         `db:"video_url"`
	SocialMedia  json.RawMessage `db:"social_media"`
	StartDate    time.Time       `db:"start_date"`
	EndDate      time.Time       `db:"end_date"`
	FormID       *uuid.UUID      `db:"form_id"`
	CreatedAt    time.Time       `db:"created_at"`
	UpdatedAt    time.Time       `db:"updated_at"`
	BrandName    string          `db:"brand_name"`
	Schema       json.RawMessage `db:"schema"`
}

// EventRepositoryImpl implements the contracts.EventRepository interface
type EventRepositoryImpl struct {
	baseRepo *repo.BaseRepository
}

// NewEventRepositoryImpl creates a new EventRepositoryImpl
// If db is nil, uses BaseRepository's default DB connection
func NewEventRepositoryImpl(db *sqlx.DB) contracts.EventRepository {
	baseRepo := repo.NewBaseRepository()
	if db != nil {
		baseRepo.DB = db
	}
	return &EventRepositoryImpl{
		baseRepo: baseRepo,
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
		// Create new event - use the domain-generated ID
		eventId := event.ID()
		_, err = tx.Exec(`
			INSERT INTO events (id, name, description, status, slug, category_id, mode_type, location, payment, duration_type, image_url, video_url, social_media, start_date, end_date, user_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`,
			eventId,
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
		)

		if err != nil {
			return err
		}

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
			    start_date = $13, end_date = $14, updated_at = NOW()
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
	query := `
		SELECT 
			e.id, e.user_id, e.name, e.description, e.slug, e.status,
			e.category_id, e.mode_type, e.location, e.payment, e.duration_type,
			e.image_url, e.video_url, e.social_media, e.start_date, e.end_date,
			e.form_id, e.created_at, e.updated_at,
			u.name as brand_name, f.schema
		FROM events e
		JOIN users u ON e.user_id = u.id
		LEFT JOIN forms f ON e.form_id = f.id
		WHERE e.id = $1`

	var row eventRowDTO

	var err error
	if tx != nil {
		err = tx.QueryRow(query, id).Scan(
			&row.EventID, &row.UserID, &row.Name, &row.Description, &row.Slug, &row.Status,
			&row.CategoryID, &row.ModeType, &row.Location, &row.Payment, &row.DurationType,
			&row.ImageURL, &row.VideoURL, &row.SocialMedia, &row.StartDate, &row.EndDate,
			&row.FormID, &row.CreatedAt, &row.UpdatedAt, &row.BrandName, &row.Schema,
		)
	} else {
		err = r.baseRepo.DB.QueryRow(query, id).Scan(
			&row.EventID, &row.UserID, &row.Name, &row.Description, &row.Slug, &row.Status,
			&row.CategoryID, &row.ModeType, &row.Location, &row.Payment, &row.DurationType,
			&row.ImageURL, &row.VideoURL, &row.SocialMedia, &row.StartDate, &row.EndDate,
			&row.FormID, &row.CreatedAt, &row.UpdatedAt, &row.BrandName, &row.Schema,
		)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return r.toDomainEventFromDTO(&row)
}

// FindByUserID finds events by user ID
func (r *EventRepositoryImpl) FindByUserID(userId uuid.UUID, query contracts.QueryParams) (contracts.EventList, error) {
	sqlString := `
		SELECT 
			e.id, e.user_id, e.name, e.description, e.slug, e.status,
		e.category_id, e.mode_type, e.location, e.payment, e.duration_type,
			e.image_url, e.video_url, e.social_media, e.start_date, e.end_date,
			e.form_id, e.created_at, e.updated_at,
			u.name as brand_name
	FROM events e 
	JOIN users u ON e.user_id = u.id
		LEFT JOIN forms f ON e.form_id = f.id
		WHERE e.user_id = :user_id
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

	params := pager.Params{
		"user_id": userId,
	}

	pager := pager.NewPager[contracts.EventItem]().SetQuery(sqlString).SetConditions(conditions).SetParams(params).SetPage(query.Page)

	if query.Size > 0 {
		pager = pager.SetSize(query.Size)
	}

	result, err := pager.Query()
	if err != nil {
		return contracts.EventList{}, err
	}

	return contracts.EventList{
		Items: result.GetItems(),
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

func (r *EventRepositoryImpl) toDomainEventFromDTO(row *eventRowDTO) (*domain.Event, error) {
	// Create value objects
	eventName, err := domain.NewEventName(row.Name)
	if err != nil {
		return nil, err
	}

	eventDescription := domain.NewEventDescription(row.Description)

	dateRange, err := domain.NewEventDateRange(row.StartDate, row.EndDate)
	if err != nil {
		return nil, err
	}

	modeTypeVO, err := domain.NewEventModeType(row.ModeType)
	if err != nil {
		return nil, err
	}

	locationValue := ""
	if row.Location != nil {
		locationValue = *row.Location
	}
	// Use NewLocationWithDefaults for reconstruction since persisted locations may have defaults
	locationVO := domain.NewLocationWithDefaults(locationValue, modeTypeVO)

	paymentVO, err := domain.NewEventPayment(row.Payment)
	if err != nil {
		return nil, err
	}

	accessType, err := domain.NewAccessType("public") // Default, should be stored in DB
	if err != nil {
		return nil, err
	}

	durationTypeVO, err := domain.NewDurationType(row.DurationType)
	if err != nil {
		return nil, err
	}

	registrationType, err := domain.NewRegistrationType("ticket") // Default, should be stored in DB
	if err != nil {
		return nil, err
	}

	statusVO, err := domain.NewEventStatus(row.Status)
	if err != nil {
		return nil, err
	}

	// This reconstructs an event from persistence without raising creation events
	return r.reconstructEvent(
		row.EventID, row.UserID, eventName, eventDescription, statusVO,
		dateRange, locationVO, paymentVO, accessType, durationTypeVO, registrationType,
		row.CategoryID, row.ImageURL, row.VideoURL, row.SocialMedia, row.Schema, row.FormID, row.BrandName, row.Slug,
		row.CreatedAt, row.UpdatedAt,
	), nil
}

func (r *EventRepositoryImpl) reconstructEvent(
	id, userId uuid.UUID, name domain.EventName, description domain.EventDescription,
	status domain.EventStatus, dateRange domain.EventDateRange, location domain.Location,
	payment domain.EventPayment, accessType domain.AccessType, durationType domain.DurationType,
	registrationType domain.RegistrationType, categoryId *int, imageURL, videoURL *string,
	socialMedia, schema json.RawMessage, formId *uuid.UUID, brandName, slug string,
	createdAt, updatedAt time.Time,
) *domain.Event {
	return domain.ReconstructEvent(
		id, userId, name, description, status, dateRange, location,
		payment, accessType, durationType, registrationType,
		categoryId, imageURL, videoURL, socialMedia, schema,
		formId, brandName, slug, createdAt, updatedAt,
	)
}

func (r *EventRepositoryImpl) createFormForEvent(tx *sql.Tx, event *domain.Event) (uuid.UUID, error) {
	// This would need to be implemented based on the form creation logic
	// For now, return a placeholder
	return uuid.New(), nil
}

// GetEventsReport retrieves events report using the database function
func (r *EventRepositoryImpl) GetEventsReport(userId uuid.UUID) (json.RawMessage, error) {
	var reportJson json.RawMessage

	err := r.baseRepo.DB.QueryRow(`
		SELECT * FROM get_events_report($1)`,
		userId,
	).Scan(&reportJson)

	if err != nil {
		return nil, fmt.Errorf("failed to get events report: %w", err)
	}

	return reportJson, nil
}
