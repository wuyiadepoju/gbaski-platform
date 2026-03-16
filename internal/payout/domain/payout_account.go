package domain

import "github.com/google/uuid"

// PayoutAccount is the aggregate root for a user's payout account
type PayoutAccount struct {
	id            int
	userID        uuid.UUID
	accountNumber string
	accountName   string
	bankName      string
	bankID        int
	payoutType    PayoutType
	feeRates      FeeRates
	currency      Currency

	// Domain events
	domainEvents []DomainEvent
}

// NewPayoutAccount creates a new PayoutAccount aggregate
// The ID is set after persistence (assigned by DB); use SetID after saving.
func NewPayoutAccount(
	userID uuid.UUID,
	accountNumber, accountName string,
	bankID int,
	currency Currency,
) (*PayoutAccount, error) {
	if userID == uuid.Nil {
		return nil, ErrEmptyUserID
	}

	account := &PayoutAccount{
		userID:        userID,
		accountNumber: accountNumber,
		accountName:   accountName,
		bankID:        bankID,
		payoutType:    PayoutTypeWeekly, // default
		feeRates:      DefaultFeeRates(),
		currency:      currency,
		domainEvents:  []DomainEvent{},
	}

	return account, nil
}

// ReconstructPayoutAccount rebuilds a PayoutAccount from persistence without raising domain events
func ReconstructPayoutAccount(
	id int,
	userID uuid.UUID,
	accountNumber, accountName, bankName string,
	bankID int,
	payoutType PayoutType,
	currency Currency,
) *PayoutAccount {
	return &PayoutAccount{
		id:            id,
		userID:        userID,
		accountNumber: accountNumber,
		accountName:   accountName,
		bankName:      bankName,
		bankID:        bankID,
		payoutType:    payoutType,
		feeRates:      DefaultFeeRates(),
		currency:      currency,
		domainEvents:  []DomainEvent{},
	}
}

// SetID sets the database-assigned id after persistence
func (a *PayoutAccount) SetID(id int) {
	a.id = id
	a.addDomainEvent(NewPayoutAccountCreatedEvent(a.id, a.userID, a.currency, a.accountNumber))
}

// ChangePayoutType changes the payout type for this account
func (a *PayoutAccount) ChangePayoutType(newType PayoutType) {
	oldType := a.payoutType
	a.payoutType = newType
	a.addDomainEvent(NewPayoutTypeChangedEvent(a.id, a.userID, oldType, newType))
}

// Getters
func (a *PayoutAccount) ID() int                { return a.id }
func (a *PayoutAccount) UserID() uuid.UUID      { return a.userID }
func (a *PayoutAccount) AccountNumber() string  { return a.accountNumber }
func (a *PayoutAccount) AccountName() string    { return a.accountName }
func (a *PayoutAccount) BankName() string       { return a.bankName }
func (a *PayoutAccount) BankID() int            { return a.bankID }
func (a *PayoutAccount) PayoutType() PayoutType { return a.payoutType }
func (a *PayoutAccount) FeeRates() FeeRates     { return a.feeRates }
func (a *PayoutAccount) Currency() Currency     { return a.currency }

// Domain events
func (a *PayoutAccount) DomainEvents() []DomainEvent {
	return a.domainEvents
}

func (a *PayoutAccount) ClearDomainEvents() {
	a.domainEvents = []DomainEvent{}
}

func (a *PayoutAccount) addDomainEvent(event DomainEvent) {
	a.domainEvents = append(a.domainEvents, event)
}
