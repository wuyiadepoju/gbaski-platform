package adapters

import (
	"github.com/gbaski/gbaski-ext/auth"
	"github.com/gbaski/gbaski-platform/internal/payout/contracts"
	"github.com/google/uuid"
)

// AuthServiceImpl wraps gbaski-ext/auth.AuthHandler
type AuthServiceImpl struct {
	handler *auth.AuthHandler
}

func NewAuthServiceImpl() contracts.AuthService {
	return &AuthServiceImpl{handler: auth.NewAuthHandler()}
}

func (a *AuthServiceImpl) GetUserByID(userID uuid.UUID) (*contracts.AuthUser, error) {
	user, err := a.handler.GetUserById(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}
	return &contracts.AuthUser{
		ID:    userID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}
