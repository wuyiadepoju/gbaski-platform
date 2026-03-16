package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Event is the aggregate root for the event domain
type Event struct {
	id               uuid.UUID
	userId           uuid.UUID
	name             EventName
	description      EventDescription
	status           EventStatus
	dateRange        EventDateRange
	location         Location
	payment          EventPayment
	accessType       AccessType
	durationType     DurationType
	registrationType RegistrationType
	categoryId       *int
	imageURL         *string
	videoURL         *string
	socialMedia      json.RawMessage
	schema           json.RawMessage
	formId           *uuid.UUID
	brandName        string
	slug             string
	createdAt        time.Time
	updatedAt        time.Time

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
		return nil, ErrEmptyUserID
	}

	status := EventStatusDraft
	// Business rule: Public free events are automatically published
	if accessType.IsPublic() && payment.IsFree() {
		status = EventStatusPublished
	}

	event := &Event{
		id:               uuid.New(),
		userId:           userId,
		name:             name,
		description:      description,
		status:           status,
		dateRange:        dateRange,
		location:         location,
		payment:          payment,
		accessType:       accessType,
		durationType:     durationType,
		registrationType: registrationType,
		socialMedia:      json.RawMessage("{}"),
		brandName:        brandName,
		slug:             name.Slug(),
		createdAt:        time.Now(),
		updatedAt:        time.Now(),
		domainEvents:     []DomainEvent{},
	}

	event.addDomainEvent(NewEventCreatedEvent(event.id, event.userId, event.name.Value()))

	return event, nil
}

// Publish publishes the event
func (e *Event) Publish() error {
	if !e.status.IsDraft() {
		return ErrEventNotDraft
	}

	oldStatus := e.status
	e.status = EventStatusPublished
	e.updatedAt = time.Now()

	e.addDomainEvent(NewEventPublishedEvent(e.id, e.userId))
	e.addDomainEvent(NewEventStatusChangedEvent(e.id, e.userId, oldStatus, e.status))

	return nil
}

// UpdateDates updates the event dates
func (e *Event) UpdateDates(newDateRange EventDateRange) error {
	if e.status.IsEnded() {
		return ErrEventEnded
	}

	oldStartDate := e.dateRange.startDate
	oldEndDate := e.dateRange.endDate

	e.dateRange = newDateRange
	e.updatedAt = time.Now()

	e.addDomainEvent(NewEventDatesChangedEvent(e.id, e.userId, oldStartDate, newDateRange.startDate, oldEndDate, newDateRange.endDate))

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

	e.addDomainEvent(NewEventStatusChangedEvent(e.id, e.userId, oldStatus, newStatus))

	return nil
}

// UpdateImage updates the event image URL
func (e *Event) UpdateImage(imageURL string) error {
	if imageURL == "" {
		return ErrEmptyImageURL
	}
	e.imageURL = &imageURL
	e.updatedAt = time.Now()
	return nil
}

// UpdateVideo updates the event video URL
func (e *Event) UpdateVideo(videoURL string) error {
	if videoURL == "" {
		return ErrEmptyVideoURL
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
func (e *Event) UpdateDescription(newDescription EventDescription) error {
	// Description can be empty, but we validate it's not nil/zero value
	// In this case, EventDescription is a value object that can be empty
	e.description = newDescription
	e.updatedAt = time.Now()
	return nil
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

// MarkAsUpdated marks the event as updated
func (e *Event) MarkAsUpdated() {
	e.updatedAt = time.Now()
	e.addDomainEvent(NewEventUpdatedEvent(e.id, e.userId))
}

// Delete marks the event for deletion
func (e *Event) Delete() {
	e.addDomainEvent(NewEventDeletedEvent(e.id, e.userId))
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

// ReconstructEvent reconstructs an event from persistence (doesn't raise creation events)
// This is a factory function used by the repository layer for reconstruction
// It should not be used for creating new events - use NewEvent instead
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
	socialMedia, schema json.RawMessage,
	formId *uuid.UUID,
	brandName, slug string,
	createdAt, updatedAt time.Time,
) *Event {
	return &Event{
		id:               id,
		userId:           userId,
		name:             name,
		description:      description,
		status:           status,
		dateRange:        dateRange,
		location:         location,
		payment:          payment,
		accessType:       accessType,
		durationType:     durationType,
		registrationType: registrationType,
		categoryId:       categoryId,
		imageURL:         imageURL,
		videoURL:         videoURL,
		socialMedia:      socialMedia,
		schema:           schema,
		formId:           formId,
		brandName:        brandName,
		slug:             slug,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
		domainEvents:     []DomainEvent{},
	}
}
