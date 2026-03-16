package domain

import "errors"

// Domain errors - Form aggregate
var (
	// ErrEmptyFormName indicates that the form name cannot be empty
	ErrEmptyFormName = errors.New("form name cannot be empty")

	// ErrInvalidPrice indicates that the form price is invalid (e.g. negative)
	ErrInvalidPrice = errors.New("form price cannot be negative")

	// ErrZeroStartDate indicates that the start date cannot be zero
	ErrZeroStartDate = errors.New("form start date cannot be zero")

	// ErrZeroEndDate indicates that the end date cannot be zero
	ErrZeroEndDate = errors.New("form end date cannot be zero")

	// ErrInvalidDateRange indicates that start date must be before or equal to end date
	ErrInvalidDateRange = errors.New("form start date must be before or equal to end date")
)
