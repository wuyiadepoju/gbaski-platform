package domain

// PayoutType represents the payout type value object
type PayoutType string

const (
	PayoutTypeInstant PayoutType = "instant"
	PayoutTypeWeekly  PayoutType = "weekly"
)

// NewPayoutType creates a validated PayoutType value object
func NewPayoutType(payoutType string) (PayoutType, error) {
	switch payoutType {
	case "instant", "weekly":
		return PayoutType(payoutType), nil
	default:
		return "", NewValidationError("payoutType", ErrInvalidPayoutType.Error()+": "+payoutType)
	}
}

func (p PayoutType) Value() string {
	return string(p)
}

func (p PayoutType) IsInstant() bool {
	return p == PayoutTypeInstant
}

func (p PayoutType) IsWeekly() bool {
	return p == PayoutTypeWeekly
}

// Currency represents a currency value object
type Currency string

const (
	CurrencyNGN Currency = "NGN"
	CurrencyUSD Currency = "USD"
)

// NewCurrency creates a validated Currency value object
func NewCurrency(currency string) (Currency, error) {
	switch currency {
	case "NGN", "USD":
		return Currency(currency), nil
	default:
		return "", NewValidationError("currency", ErrInvalidCurrency.Error()+": "+currency)
	}
}

func (c Currency) Value() string {
	return string(c)
}

// FeeRates represents the payout fee rates value object
type FeeRates struct {
	Instant int
	Weekly  int
}

// DefaultFeeRates returns the default fee rates
func DefaultFeeRates() FeeRates {
	return FeeRates{
		Instant: 2,
		Weekly:  0,
	}
}

func (f FeeRates) InstantRate() int {
	return f.Instant
}

func (f FeeRates) WeeklyRate() int {
	return f.Weekly
}

// ProviderState represents the state of a payment provider integration
type ProviderState string

const (
	ProviderStateInitiated ProviderState = "initiated"
	ProviderStateCompleted ProviderState = "completed"
)

// NewProviderState creates a validated ProviderState value object
func NewProviderState(state string) (ProviderState, error) {
	switch state {
	case "initiated", "completed":
		return ProviderState(state), nil
	default:
		return "", NewValidationError("state", ErrInvalidProviderState.Error()+": "+state)
	}
}

func (p ProviderState) Value() string {
	return string(p)
}

func (p ProviderState) IsInitiated() bool {
	return p == ProviderStateInitiated
}

func (p ProviderState) IsCompleted() bool {
	return p == ProviderStateCompleted
}
