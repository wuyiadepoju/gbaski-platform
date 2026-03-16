package domain

import "errors"

// Domain errors - PayoutAccount aggregate
var (
	// ErrEmptyUserID indicates that the user ID cannot be empty
	ErrEmptyUserID = errors.New("user ID cannot be empty")

	// ErrPayoutAccountNotFound indicates that the payout account was not found
	ErrPayoutAccountNotFound = errors.New("payout account not found")

	// ErrInvalidPayoutType indicates that the payout type is invalid
	ErrInvalidPayoutType = errors.New("invalid payout type, must be 'instant' or 'weekly'")

	// ErrInvalidCurrency indicates that the currency is invalid
	ErrInvalidCurrency = errors.New("invalid currency, must be 'NGN' or 'USD'")
)

// Domain errors - PaymentProvider aggregate
var (
	// ErrPaymentProviderNotFound indicates that the payment provider was not found
	ErrPaymentProviderNotFound = errors.New("payment provider not found")

	// ErrInvalidProviderState indicates that the provider state is invalid
	ErrInvalidProviderState = errors.New("invalid provider state, must be 'initiated' or 'completed'")

	// ErrEmptyProviderName indicates that the provider name cannot be empty
	ErrEmptyProviderName = errors.New("provider name cannot be empty")

	// ErrEmptyPublicKey indicates that the public key cannot be empty
	ErrEmptyPublicKey = errors.New("public key cannot be empty")

	// ErrEmptySecretKey indicates that the secret key cannot be empty
	ErrEmptySecretKey = errors.New("secret key cannot be empty")
)

// ValidationError represents a validation error with a field name
type ValidationError struct {
	Field   string
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return e.Field + ": " + e.Message
	}
	return e.Message
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// IsValidationError checks if an error is a ValidationError
func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}
