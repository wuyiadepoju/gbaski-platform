package domain

import (
	"strings"
	"time"

	"github.com/gbaski/gbaski-shared/util"
)

// EventName represents an event name value object
type EventName struct {
	value string
}

func NewEventName(name string) (EventName, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return EventName{}, ErrEmptyEventName
	}
	if len(name) < 3 {
		return EventName{}, ErrEventNameTooShort
	}
	if len(name) > 200 {
		return EventName{}, ErrEventNameTooLong
	}
	return EventName{value: name}, nil
}

func (n EventName) Value() string {
	return n.value
}

func (n EventName) Slug() string {
	return util.Slugify(n.value)
}

// EventDescription represents an event description value object
type EventDescription struct {
	value string
}

func NewEventDescription(description string) EventDescription {
	return EventDescription{value: strings.TrimSpace(description)}
}

func (d EventDescription) Value() string {
	return d.value
}

// EventDateRange represents a date range for an event
type EventDateRange struct {
	startDate time.Time
	endDate   time.Time
}

func NewEventDateRange(startDate, endDate time.Time) (EventDateRange, error) {
	if startDate.IsZero() {
		return EventDateRange{}, ErrZeroStartDate
	}
	if endDate.IsZero() {
		return EventDateRange{}, ErrZeroEndDate
	}
	if startDate.After(endDate) {
		return EventDateRange{}, ErrInvalidDateRange
	}
	return EventDateRange{
		startDate: startDate,
		endDate:   endDate,
	}, nil
}

func (dr EventDateRange) StartDate() time.Time {
	return dr.startDate
}

func (dr EventDateRange) EndDate() time.Time {
	return dr.endDate
}

func (dr EventDateRange) Duration() time.Duration {
	return dr.endDate.Sub(dr.startDate)
}

func (dr EventDateRange) IsInPast() bool {
	return dr.endDate.Before(time.Now())
}

func (dr EventDateRange) IsInFuture() bool {
	return dr.startDate.After(time.Now())
}

func (dr EventDateRange) IsActive() bool {
	now := time.Now()
	return !dr.startDate.After(now) && !dr.endDate.Before(now)
}

// Location represents an event location
type Location struct {
	value    string
	modeType EventModeType
}

// NewLocation creates a new Location value object
// If location is empty, returns an error - default values should be handled by application layer
func NewLocation(location string, modeType EventModeType) (Location, error) {
	location = strings.TrimSpace(location)
	if location == "" {
		return Location{}, ErrEmptyLocation
	}
	return Location{
		value:    location,
		modeType: modeType,
	}, nil
}

// NewLocationWithDefaults creates a Location with default values for empty location
// This is a convenience method for cases where defaults are acceptable
func NewLocationWithDefaults(location string, modeType EventModeType) Location {
	location = strings.TrimSpace(location)
	if location == "" {
		if modeType == EventModePhysical {
			location = "Undisclosed Location"
		} else {
			location = "http://virtual"
		}
	}
	return Location{
		value:    location,
		modeType: modeType,
	}
}

func (l Location) Value() string {
	return l.value
}

func (l Location) ModeType() EventModeType {
	return l.modeType
}

// EventModeType represents the mode type of an event
type EventModeType string

const (
	EventModePhysical EventModeType = "physical"
	EventModeVirtual  EventModeType = "virtual"
)

func NewEventModeType(modeType string) (EventModeType, error) {
	switch modeType {
	case "physical", "virtual":
		return EventModeType(modeType), nil
	default:
		return "", NewValidationError("modeType", ErrInvalidEventModeType.Error()+": "+modeType)
	}
}

// EventStatus represents the status of an event
type EventStatus string

const (
	EventStatusDraft     EventStatus = "draft"
	EventStatusPublished EventStatus = "published"
	EventStatusEnded     EventStatus = "ended"
)

func NewEventStatus(status string) (EventStatus, error) {
	switch status {
	case "draft", "published", "ended":
		return EventStatus(status), nil
	default:
		return "", NewValidationError("status", ErrInvalidEventStatus.Error()+": "+status)
	}
}

func (s EventStatus) Value() string {
	return string(s)
}

func (s EventStatus) IsDraft() bool {
	return s == EventStatusDraft
}

func (s EventStatus) IsPublished() bool {
	return s == EventStatusPublished
}

func (s EventStatus) IsEnded() bool {
	return s == EventStatusEnded
}

// EventPayment represents the payment type of an event
type EventPayment string

const (
	EventPaymentFree EventPayment = "free"
	EventPaymentPaid EventPayment = "paid"
)

func NewEventPayment(payment string) (EventPayment, error) {
	switch payment {
	case "free", "paid":
		return EventPayment(payment), nil
	default:
		return "", NewValidationError("payment", ErrInvalidEventPayment.Error()+": "+payment)
	}
}

func (p EventPayment) Value() string {
	return string(p)
}

func (p EventPayment) IsFree() bool {
	return p == EventPaymentFree
}

// AccessType represents the access type of an event
type AccessType string

const (
	AccessTypePublic  AccessType = "public"
	AccessTypePrivate AccessType = "private"
)

func NewAccessType(accessType string) (AccessType, error) {
	switch accessType {
	case "public", "private":
		return AccessType(accessType), nil
	default:
		return "", NewValidationError("accessType", ErrInvalidAccessType.Error()+": "+accessType)
	}
}

func (a AccessType) Value() string {
	return string(a)
}

func (a AccessType) IsPublic() bool {
	return a == AccessTypePublic
}

// DurationType represents the duration type of an event
type DurationType string

const (
	DurationTypeSingle DurationType = "single"
	DurationTypeMulti  DurationType = "multi"
)

func NewDurationType(durationType string) (DurationType, error) {
	switch durationType {
	case "single", "multi":
		return DurationType(durationType), nil
	default:
		return "", NewValidationError("durationType", ErrInvalidDurationType.Error()+": "+durationType)
	}
}

// RegistrationType represents the registration type
type RegistrationType string

const (
	RegistrationTypeTicket RegistrationType = "ticket"
	RegistrationTypeEntry  RegistrationType = "entry"
)

func NewRegistrationType(regType string) (RegistrationType, error) {
	switch regType {
	case "ticket", "entry":
		return RegistrationType(regType), nil
	default:
		return "", NewValidationError("registrationType", ErrInvalidRegistrationType.Error()+": "+regType)
	}
}

// Currency represents a currency type
type Currency string

const (
	CurrencyNGN Currency = "NGN"
	CurrencyUSD Currency = "USD"
)

// FeePayer represents who pays the fee
type FeePayer string

const (
	FeePayerBuyer  FeePayer = "buyer"
	FeePayerSeller FeePayer = "seller"
	FeePayerWallet FeePayer = "wallet"
)

// TicketPayment represents ticket payment type
type TicketPayment string

const (
	TicketPaymentFree TicketPayment = "free"
	TicketPaymentPaid TicketPayment = "paid"
)

// MediaType represents the type of media
type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
	MediaTypeAudio MediaType = "audio"
	MediaTypeFile  MediaType = "file"
	MediaTypeOther MediaType = "other"
)

// EventConfig represents event configuration
type EventConfig struct {
	Appearance string            `json:"appearance"`
	Social     map[string]string `json:"social"`
}
