package domain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Event is the aggregate root for the event domain
type Event struct {
	id              uuid.UUID
	userId          uuid.UUID
	name            EventName
	description     EventDescription
	status          EventStatus
	dateRange       EventDateRange
	location        Location
	payment         EventPayment
	accessType      AccessType
	durationType    DurationType
	registrationType RegistrationType
	categoryId      *int
	imageURL        *string
	videoURL        *string
	socialMedia     json.RawMessage
	configs         json.RawMessage
	schema          json.RawMessage
	formId          *uuid.UUID
	brandName       string
	slug            string
	createdAt       time.Time
	updatedAt       time.Time

	// Domain events
	domainEvents []DomainEvent
}

// NewEvent creates a new event aggregate
func NewEvent(
	userId uuid.UUID,
	name EventName,
	description EventDescription,
	dateRange EventDateRange,
	location Location,
	payment EventPayment,
	accessType AccessType,
	durationType DurationType,
	registrationType RegistrationType,
	brandName string,
) (*Event, error) {
	if userId == uuid.Nil {
		return nil, errors.New("user ID cannot be empty")
	}

	status := EventStatusDraft
	// Business rule: Public free events are automatically published
	if accessType.IsPublic() && payment.IsFree() {
		status = EventStatusPublished
	}

	event := &Event{
		id:              uuid.New(),
		userId:          userId,
		name:            name,
		description:     description,
		status:          status,
		dateRange:       dateRange,
		location:        location,
		payment:         payment,
		accessType:      accessType,
		durationType:    durationType,
		registrationType: registrationType,
		socialMedia:     json.RawMessage("{}"),
		configs:         json.RawMessage("{}"),
		brandName:       brandName,
		slug:            name.Slug(),
		createdAt:       time.Now(),
		updatedAt:       time.Now(),
		domainEvents:   []DomainEvent{},
	}

	event.addDomainEvent(EventCreatedEvent{
		EventID:    event.id,
		UserID:     event.userId,
		EventName:  event.name.Value(),
		OccurredAt: time.Now(),
	})

	return event, nil
}

// Publish publishes the event
func (e *Event) Publish() error {
	if !e.status.IsDraft() {
		return errors.New("only draft events can be published")
	}

	oldStatus := e.status
	e.status = EventStatusPublished
	e.updatedAt = time.Now()

	e.addDomainEvent(EventPublishedEvent{
		EventID:    e.id,
		UserID:     e.userId,
		OccurredAt: time.Now(),
	})

	e.addDomainEvent(EventStatusChangedEvent{
		EventID:    e.id,
		UserID:     e.userId,
		OldStatus:  oldStatus,
		NewStatus:  e.status,
		OccurredAt: time.Now(),
	})

	return nil
}

// UpdateDates updates the event dates
func (e *Event) UpdateDates(newDateRange EventDateRange) error {
	if e.status.IsEnded() {
		return errors.New("cannot update dates for ended events")
	}

	oldStartDate := e.dateRange.startDate
	oldEndDate := e.dateRange.endDate

	e.dateRange = newDateRange
	e.updatedAt = time.Now()

	e.addDomainEvent(EventDatesChangedEvent{
		EventID:      e.id,
		UserID:       e.userId,
		OldStartDate: oldStartDate,
		NewStartDate: newDateRange.startDate,
		OldEndDate:   oldEndDate,
		NewEndDate:   newDateRange.endDate,
		OccurredAt:   time.Now(),
	})

	return nil
}

// UpdateStatus updates the event status
func (e *Event) UpdateStatus(newStatus EventStatus) error {
	if e.status == newStatus {
		return nil // No change
	}

	oldStatus := e.status
	e.status = newStatus
	e.updatedAt = time.Now()

	e.addDomainEvent(EventStatusChangedEvent{
		EventID:    e.id,
		UserID:     e.userId,
		OldStatus:  oldStatus,
		NewStatus:  newStatus,
		OccurredAt: time.Now(),
	})

	return nil
}

// UpdateImage updates the event image URL
func (e *Event) UpdateImage(imageURL string) error {
	if imageURL == "" {
		return errors.New("image URL cannot be empty")
	}
	e.imageURL = &imageURL
	e.updatedAt = time.Now()
	return nil
}

// UpdateVideo updates the event video URL
func (e *Event) UpdateVideo(videoURL string) error {
	if videoURL == "" {
		return errors.New("video URL cannot be empty")
	}
	e.videoURL = &videoURL
	e.updatedAt = time.Now()
	return nil
}

// UpdateName updates the event name
func (e *Event) UpdateName(newName EventName) error {
	e.name = newName
	e.slug = newName.Slug()
	e.updatedAt = time.Now()
	return nil
}

// UpdateDescription updates the event description
func (e *Event) UpdateDescription(newDescription EventDescription) {
	e.description = newDescription
	e.updatedAt = time.Now()
}

// SetFormId sets the form ID for the event
func (e *Event) SetFormId(formId uuid.UUID) {
	e.formId = &formId
	e.updatedAt = time.Now()
}

// SetCategoryId sets the category ID
func (e *Event) SetCategoryId(categoryId int) {
	e.categoryId = &categoryId
	e.updatedAt = time.Now()
}

// SetSchema sets the schema for the event
func (e *Event) SetSchema(schema json.RawMessage) {
	e.schema = schema
	e.updatedAt = time.Now()
}

// SetConfigs sets the configs for the event
func (e *Event) SetConfigs(configs json.RawMessage) {
	e.configs = configs
	e.updatedAt = time.Now()
}

// MarkAsUpdated marks the event as updated
func (e *Event) MarkAsUpdated() {
	e.updatedAt = time.Now()
	e.addDomainEvent(EventUpdatedEvent{
		EventID:    e.id,
		UserID:     e.userId,
		OccurredAt: time.Now(),
	})
}

// Delete marks the event for deletion
func (e *Event) Delete() {
	e.addDomainEvent(EventDeletedEvent{
		EventID:    e.id,
		UserID:     e.userId,
		OccurredAt: time.Now(),
	})
}

// Getters
func (e *Event) ID() uuid.UUID {
	return e.id
}

func (e *Event) UserID() uuid.UUID {
	return e.userId
}

func (e *Event) Name() EventName {
	return e.name
}

func (e *Event) Description() EventDescription {
	return e.description
}

func (e *Event) Status() EventStatus {
	return e.status
}

func (e *Event) DateRange() EventDateRange {
	return e.dateRange
}

func (e *Event) Location() Location {
	return e.location
}

func (e *Event) Payment() EventPayment {
	return e.payment
}

func (e *Event) AccessType() AccessType {
	return e.accessType
}

func (e *Event) DurationType() DurationType {
	return e.durationType
}

func (e *Event) RegistrationType() RegistrationType {
	return e.registrationType
}

func (e *Event) CategoryId() *int {
	return e.categoryId
}

func (e *Event) ImageURL() *string {
	return e.imageURL
}

func (e *Event) VideoURL() *string {
	return e.videoURL
}

func (e *Event) SocialMedia() json.RawMessage {
	return e.socialMedia
}

func (e *Event) Configs() json.RawMessage {
	return e.configs
}

func (e *Event) Schema() json.RawMessage {
	return e.schema
}

func (e *Event) FormId() *uuid.UUID {
	return e.formId
}

func (e *Event) BrandName() string {
	return e.brandName
}

func (e *Event) Slug() string {
	return e.slug
}

func (e *Event) CreatedAt() time.Time {
	return e.createdAt
}

func (e *Event) UpdatedAt() time.Time {
	return e.updatedAt
}

// Domain events
func (e *Event) DomainEvents() []DomainEvent {
	return e.domainEvents
}

func (e *Event) ClearDomainEvents() {
	e.domainEvents = []DomainEvent{}
}

func (e *Event) addDomainEvent(event DomainEvent) {
	e.domainEvents = append(e.domainEvents, event)
}

// SetID is used by repository to set the ID after persistence
func (e *Event) SetID(id uuid.UUID) {
	e.id = id
}

// SetCreatedAt is used by repository to set creation time
func (e *Event) SetCreatedAt(createdAt time.Time) {
	e.createdAt = createdAt
}

// ReconstructEvent reconstructs an event from persistence (doesn't raise creation events)
func ReconstructEvent(
	id, userId uuid.UUID,
	name EventName,
	description EventDescription,
	status EventStatus,
	dateRange EventDateRange,
	location Location,
	payment EventPayment,
	accessType AccessType,
	durationType DurationType,
	registrationType RegistrationType,
	categoryId *int,
	imageURL, videoURL *string,
	socialMedia, configs, schema json.RawMessage,
	formId *uuid.UUID,
	brandName, slug string,
	createdAt, updatedAt time.Time,
) *Event {
	return &Event{
		id:              id,
		userId:          userId,
		name:            name,
		description:     description,
		status:          status,
		dateRange:       dateRange,
		location:        location,
		payment:         payment,
		accessType:      accessType,
		durationType:    durationType,
		registrationType: registrationType,
		categoryId:      categoryId,
		imageURL:        imageURL,
		videoURL:        videoURL,
		socialMedia:     socialMedia,
		configs:         configs,
		schema:          schema,
		formId:          formId,
		brandName:       brandName,
		slug:            slug,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
		domainEvents:    []DomainEvent{},
	}
}
