package domain

import "github.com/google/uuid"

// PaymentProvider is an aggregate representing a seller's payment provider integration
type PaymentProvider struct {
	userID    uuid.UUID
	provider  string
	publicKey string
	secretKey string
	balance   any // *wallet.WalletBalance — stored as any to avoid domain → infra dependency

	// Domain events
	domainEvents []DomainEvent
}

// NewPaymentProvider creates a new PaymentProvider aggregate
func NewPaymentProvider(
	userID uuid.UUID,
	provider, publicKey, secretKey string,
) (*PaymentProvider, error) {
	if userID == uuid.Nil {
		return nil, ErrEmptyUserID
	}
	if provider == "" {
		return nil, ErrEmptyProviderName
	}
	if publicKey == "" {
		return nil, ErrEmptyPublicKey
	}
	if secretKey == "" {
		return nil, ErrEmptySecretKey
	}

	return &PaymentProvider{
		userID:       userID,
		provider:     provider,
		publicKey:    publicKey,
		secretKey:    secretKey,
		domainEvents: []DomainEvent{},
	}, nil
}

// ReconstructPaymentProvider rebuilds a PaymentProvider from persistence without raising domain events
func ReconstructPaymentProvider(
	userID uuid.UUID,
	provider, publicKey, secretKey string,
) *PaymentProvider {
	return &PaymentProvider{
		userID:       userID,
		provider:     provider,
		publicKey:    publicKey,
		secretKey:    secretKey,
		domainEvents: []DomainEvent{},
	}
}

// CompleteIntegration marks the provider as fully linked and raises a domain event
func (p *PaymentProvider) CompleteIntegration() {
	p.addDomainEvent(NewPaymentProviderLinkedEvent(p.userID, p.provider))
}

// RedactSecretKey zeroes out the secret key (for read operations)
func (p *PaymentProvider) RedactSecretKey() {
	p.secretKey = ""
}

// SetBalance enriches the provider with a wallet balance (for read projections).
// Accepts any value — typically *wallet.WalletBalance from the application layer.
func (p *PaymentProvider) SetBalance(balance any) {
	p.balance = balance
}

// Getters
func (p *PaymentProvider) UserID() uuid.UUID { return p.userID }
func (p *PaymentProvider) Provider() string  { return p.provider }
func (p *PaymentProvider) PublicKey() string { return p.publicKey }
func (p *PaymentProvider) SecretKey() string { return p.secretKey }
func (p *PaymentProvider) Balance() any      { return p.balance }

// Domain events
func (p *PaymentProvider) DomainEvents() []DomainEvent {
	return p.domainEvents
}

func (p *PaymentProvider) ClearDomainEvents() {
	p.domainEvents = []DomainEvent{}
}

func (p *PaymentProvider) addDomainEvent(event DomainEvent) {
	p.domainEvents = append(p.domainEvents, event)
}
