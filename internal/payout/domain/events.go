package domain

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents a domain event
type DomainEvent interface {
	EventType() string
	OccurredAt() time.Time
}

// PayoutAccountCreatedEvent is raised when a payout account is created or updated
type PayoutAccountCreatedEvent struct {
	accountID     int
	userID        uuid.UUID
	currency      Currency
	accountNumber string
	occurredAt    time.Time
}

func NewPayoutAccountCreatedEvent(accountID int, userID uuid.UUID, currency Currency, accountNumber string) PayoutAccountCreatedEvent {
	return PayoutAccountCreatedEvent{
		accountID:     accountID,
		userID:        userID,
		currency:      currency,
		accountNumber: accountNumber,
		occurredAt:    time.Now(),
	}
}

func (e PayoutAccountCreatedEvent) EventType() string    { return "payout.account.created" }
func (e PayoutAccountCreatedEvent) OccurredAt() time.Time { return e.occurredAt }
func (e PayoutAccountCreatedEvent) AccountID() int        { return e.accountID }
func (e PayoutAccountCreatedEvent) UserID() uuid.UUID     { return e.userID }
func (e PayoutAccountCreatedEvent) Currency() Currency    { return e.currency }
func (e PayoutAccountCreatedEvent) AccountNumber() string { return e.accountNumber }

// PayoutTypeChangedEvent is raised when the payout type is changed
type PayoutTypeChangedEvent struct {
	accountID  int
	userID     uuid.UUID
	oldType    PayoutType
	newType    PayoutType
	occurredAt time.Time
}

func NewPayoutTypeChangedEvent(accountID int, userID uuid.UUID, oldType, newType PayoutType) PayoutTypeChangedEvent {
	return PayoutTypeChangedEvent{
		accountID:  accountID,
		userID:     userID,
		oldType:    oldType,
		newType:    newType,
		occurredAt: time.Now(),
	}
}

func (e PayoutTypeChangedEvent) EventType() string    { return "payout.type.changed" }
func (e PayoutTypeChangedEvent) OccurredAt() time.Time { return e.occurredAt }
func (e PayoutTypeChangedEvent) AccountID() int        { return e.accountID }
func (e PayoutTypeChangedEvent) UserID() uuid.UUID     { return e.userID }
func (e PayoutTypeChangedEvent) OldType() PayoutType    { return e.oldType }
func (e PayoutTypeChangedEvent) NewType() PayoutType    { return e.newType }

// PaymentProviderLinkedEvent is raised when a payment provider integration is completed
type PaymentProviderLinkedEvent struct {
	userID     uuid.UUID
	provider   string
	occurredAt time.Time
}

func NewPaymentProviderLinkedEvent(userID uuid.UUID, provider string) PaymentProviderLinkedEvent {
	return PaymentProviderLinkedEvent{
		userID:     userID,
		provider:   provider,
		occurredAt: time.Now(),
	}
}

func (e PaymentProviderLinkedEvent) EventType() string    { return "payout.provider.linked" }
func (e PaymentProviderLinkedEvent) OccurredAt() time.Time { return e.occurredAt }
func (e PaymentProviderLinkedEvent) UserID() uuid.UUID     { return e.userID }
func (e PaymentProviderLinkedEvent) Provider() string      { return e.provider }
