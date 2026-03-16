package domain

import "errors"

// Domain errors - Event aggregate
var (
	// ErrEmptyUserID indicates that the user ID cannot be empty
	ErrEmptyUserID = errors.New("user ID cannot be empty")

	// ErrEventNotDraft indicates that only draft events can be published
	ErrEventNotDraft = errors.New("only draft events can be published")

	// ErrEventEnded indicates that ended events cannot be modified
	ErrEventEnded = errors.New("cannot update dates for ended events")

	// ErrEmptyImageURL indicates that image URL cannot be empty
	ErrEmptyImageURL = errors.New("image URL cannot be empty")

	// ErrEmptyVideoURL indicates that video URL cannot be empty
	ErrEmptyVideoURL = errors.New("video URL cannot be empty")
)

// Domain errors - Value objects
var (
	// ErrEmptyEventName indicates that event name cannot be empty
	ErrEmptyEventName = errors.New("event name cannot be empty")

	// ErrEventNameTooShort indicates that event name is too short
	ErrEventNameTooShort = errors.New("event name must be at least 3 characters")

	// ErrEventNameTooLong indicates that event name exceeds maximum length
	ErrEventNameTooLong = errors.New("event name cannot exceed 200 characters")

	// ErrZeroStartDate indicates that start date cannot be zero
	ErrZeroStartDate = errors.New("start date cannot be zero")

	// ErrZeroEndDate indicates that end date cannot be zero
	ErrZeroEndDate = errors.New("end date cannot be zero")

	// ErrInvalidDateRange indicates that start date must be before or equal to end date
	ErrInvalidDateRange = errors.New("start date must be before or equal to end date")

	// ErrInvalidEventModeType indicates that the event mode type is invalid
	ErrInvalidEventModeType = errors.New("invalid event mode type")

	// ErrInvalidEventStatus indicates that the event status is invalid
	ErrInvalidEventStatus = errors.New("invalid event status")

	// ErrInvalidEventPayment indicates that the event payment type is invalid
	ErrInvalidEventPayment = errors.New("invalid event payment type")

	// ErrInvalidAccessType indicates that the access type is invalid
	ErrInvalidAccessType = errors.New("invalid access type")

	// ErrInvalidDurationType indicates that the duration type is invalid
	ErrInvalidDurationType = errors.New("invalid duration type")

	// ErrInvalidRegistrationType indicates that the registration type is invalid
	ErrInvalidRegistrationType = errors.New("invalid registration type")

	// ErrEmptyLocation indicates that location cannot be empty
	ErrEmptyLocation = errors.New("location cannot be empty")

	// ErrEmptyCategoryName indicates that category name cannot be empty
	ErrEmptyCategoryName = errors.New("category name cannot be empty")

	// ErrCategoryNameTooLong indicates that category name exceeds maximum length
	ErrCategoryNameTooLong = errors.New("category name cannot exceed 200 characters")

	// ErrEmptyEventType indicates that event type cannot be empty
	ErrEmptyEventType = errors.New("event type cannot be empty")
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

// NewValidationErrorWithCause creates a new validation error with a cause
func NewValidationErrorWithCause(field, message string, err error) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Err:     err,
	}
}

// IsValidationError checks if an error is a ValidationError
func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}

// GetValidationError extracts a ValidationError from an error
func GetValidationError(err error) *ValidationError {
	var validationErr *ValidationError
	if errors.As(err, &validationErr) {
		return validationErr
	}
	return nil
}
