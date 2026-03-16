package domain

import "errors"

var (
	ErrRegistrationNotFound = errors.New("registration not found")
	ErrAttendeeNotFound     = errors.New("attendee not found")
	ErrAlreadyCheckedIn     = errors.New("already checked in")
	ErrInvalidRegStatus     = errors.New("invalid registration status")
	ErrEmptyFirstName       = errors.New("first name cannot be empty")
	ErrEmptyEmail           = errors.New("email cannot be empty")
	ErrAgentNotFound        = errors.New("agent not found")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
