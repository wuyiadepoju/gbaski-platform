package event

import (
	"github.com/gbaski/gbaski-event/pkg/event"
	"github.com/gbaski/gbaski-ext/common"
	"github.com/google/uuid"

	"github.com/gbaski/gbaski-ext/auth"
	"github.com/gbaski/gbaski-shared/util"
)

type Response = common.Response
type User = auth.User

var stringPtr = util.Ptr[string]
var boolPtr = util.Ptr[bool]

type EventCategory = event.EventCategory

type EventStatus = event.EventStatus

var (
	EventStatusDraft     = event.EventStatusDraft
	EventStatusPublished = event.EventStatusPublished
	EventStatusEnded     = event.EventStatusEnded
)

type EventPayment = event.EventPayment

type EventModeType = event.EventModeType

type RegistrationType = event.RegistrationType

type DurationType = event.DurationType

type AccessType = event.AccessType

type Currency = event.Currency

type FeePayer = event.FeePayer

var (
	FeePayerBuyer  = event.FeePayerBuyer
	FeePayerSeller = event.FeePayerSeller
	FeePayerWallet = event.FeePayerWallet
)

type EventModel = event.EventModel

type EventItem = event.EventItem

type CreateEventRequest = event.CreateEventRequest

type CreateEventResponse = event.CreateEventResponse

type EventForm = event.EventForm

type FormModel = event.FormModel

type FormItem = event.FormItem

type FormRequest = event.FormRequest

type EventQueryModel = event.EventQueryModel

type EventList = event.EventList

var (
	RegistrationTypeTicket = event.RegistrationTypeTicket
	RegistrationTypeEntry  = event.RegistrationTypeEntry
)

var (
	DurationTypeSingle = event.DurationTypeSingle
	DurationTypeMulti  = event.DurationTypeMulti
)

var (
	EventModePhysical = event.EventModePhysical
	EventModeVirtual  = event.EventModeVirtual
)

var (
	EventPaymentFree = event.EventPaymentFree
	EventPaymentPaid = event.EventPaymentPaid
)

var (
	AccessTypePublic  = event.AccessTypePublic
	AccessTypePrivate = event.AccessTypePrivate
)

var (
	CurrencyNGN = event.CurrencyNGN
	CurrencyUSD = event.CurrencyUSD
)

type EventConfig = event.EventConfig

type SchemaTicketItem = event.SchemaTicketItem

type TicketPayment = event.TicketPayment

var (
	TicketPaymentFree = event.TicketPaymentFree
	TicketPaymentPaid = event.TicketPaymentPaid
)

type MediaType = event.MediaType

var (
	MediaTypeImage = event.MediaTypeImage
	MediaTypeVideo = event.MediaTypeVideo
	MediaTypeAudio = event.MediaTypeAudio
)

type EventDetail struct {
	Item     EventItem      `json:"item"`
	Features []EventFeature `json:"features"`
}

type EventFeature struct {
	Id          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Tag         string    `json:"tag" db:"tag"`
	AccessType  string    `json:"accessType" db:"access_type"`
	FormId      uuid.UUID `json:"formId" db:"form_id"`
}

type EventStatsItem struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Currency    string  `json:"currency"`
	Quantity    int     `json:"quantity"`
	Amount      float64 `json:"amount"`
}

type EventSummary struct {
	EventID   string `json:"event_id"`
	EventName string `json:"event_name"`
	EndDate   string `json:"end_date"`
}

type UpcomingEventSummary struct {
	UpcomingEventsCount int    `json:"upcoming_events_count"`
	NextEventDate       string `json:"next_event_date"`
	NextEventID         string `json:"next_event_id"`
}

type EventsReport struct {
	CurrentActiveEvent *EventSummary         `json:"current_active_event"`
	MostRecentEvent    *EventSummary         `json:"most_recent_event"`
	UpcomingEvents     *UpcomingEventSummary `json:"upcoming_events"`
}
