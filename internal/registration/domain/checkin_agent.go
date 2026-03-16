package domain

import "github.com/google/uuid"

type CheckInAgent struct {
	ID           uuid.UUID
	Name         string
	FormID       uuid.UUID
	RegType      string
	TotalChecked int
}

func NewCheckInAgent(id uuid.UUID, name string, formID uuid.UUID, regType string) *CheckInAgent {
	return &CheckInAgent{
		ID:      id,
		Name:    name,
		FormID:  formID,
		RegType: regType,
	}
}
