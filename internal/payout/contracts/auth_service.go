package contracts

import "github.com/google/uuid"

// AuthUser represents a user returned by the auth service
type AuthUser struct {
	ID    uuid.UUID
	Name  string
	Email string
}

// AuthService defines the contract for user lookup via the auth service
type AuthService interface {
	GetUserByID(userID uuid.UUID) (*AuthUser, error)
}
