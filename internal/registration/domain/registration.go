package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Registration struct {
	ID            *int
	FormID        uuid.UUID
	BuyerID       uuid.UUID
	RegRef        string
	RegDesc       *string
	RegStatus     RegistrationStatus
	TransactionID *int
	Metadata      *json.RawMessage
	CreatedAt     time.Time
	UpdatedAt     time.Time

	// Domain events
	domainEvents []DomainEvent
}

func NewRegistration(formID, buyerID uuid.UUID, regRef string, status RegistrationStatus) (*Registration, error) {
	if regRef == "" {
		return nil, &ValidationError{Field: "RegRef", Message: "Registration reference cannot be empty"}
	}

	return &Registration{
		FormID:       formID,
		BuyerID:      buyerID,
		RegRef:       regRef,
		RegStatus:    status,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		domainEvents: []DomainEvent{},
	}, nil
}

func (r *Registration) Reconstruct(id *int, formID, buyerID uuid.UUID, regRef string, status RegistrationStatus, txID *int, metadata *json.RawMessage, createdAt, updatedAt time.Time) {
	r.ID = id
	r.FormID = formID
	r.BuyerID = buyerID
	r.RegRef = regRef
	r.RegStatus = status
	r.TransactionID = txID
	r.Metadata = metadata
	r.CreatedAt = createdAt
	r.UpdatedAt = updatedAt
}

func (r *Registration) Cancel() {
	r.RegStatus = RegistrationStatusCancelled
	r.UpdatedAt = time.Now()
}

func (r *Registration) CheckIn() {
	r.RegStatus = RegistrationStatusCheckedIn
	r.UpdatedAt = time.Now()
}

func (r *Registration) DomainEvents() []DomainEvent {
	return r.domainEvents
}

func (r *Registration) ClearDomainEvents() {
	r.domainEvents = []DomainEvent{}
}

func (r *Registration) addDomainEvent(event DomainEvent) {
	r.domainEvents = append(r.domainEvents, event)
}
