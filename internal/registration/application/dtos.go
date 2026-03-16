package application

import (
	"time"

	"github.com/gbaski/gbaski-platform/internal/registration/domain"
	"github.com/google/uuid"
)

type RegistrationDTO struct {
	ID        *int      `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  *string   `json:"lastName"`
	Email     string    `json:"email"`
	RegRef    string    `json:"regRef"`
	CreatedAt time.Time `json:"date"`
}

type AttendeeDTO struct {
	ID         int64      `json:"id"`
	FirstName  string     `json:"firstName"`
	LastName   *string    `json:"lastName"`
	Email      string     `json:"email"`
	Phone      string     `json:"phone"`
	RegRef     string     `json:"regRef"`
	TicketName string     `json:"ticketName"`
	CheckedIn  bool       `json:"checkedIn"`
	CheckedAt  *time.Time `json:"checkedAt"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type RegistrationStatsDTO struct {
	TotalRegistered int `json:"totalRegistered"`
	TotalCheckedIn  int `json:"totalCheckedIn"`
	TotalCancelled  int `json:"totalCancelled"`
}

type AgentDTO struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	TotalChecked int       `json:"totalChecked"`
	FormID       uuid.UUID `json:"formId"`
	RegType      string    `json:"regType"`
}

func ToRegistrationDTO(item domain.RegistrationItem) RegistrationDTO {
	return RegistrationDTO{
		ID:        item.ID,
		FirstName: item.FirstName,
		LastName:  item.LastName,
		Email:     item.Email,
		RegRef:    item.RegRef,
		CreatedAt: item.CreatedAt,
	}
}

func ToAttendeeDTO(a domain.Attendee) AttendeeDTO {
	return AttendeeDTO{
		ID:        a.ID,
		FirstName: a.FirstName,
		LastName:  a.LastName,
		Email:     a.Email,
		Phone:     a.Phone,
		RegRef:    a.RegRef,
		CheckedIn: a.CheckedIn,
		CheckedAt: a.CheckedAt,
		CreatedAt: a.CreatedAt,
	}
}

func ToAgentDTO(a domain.CheckInAgent) AgentDTO {
	return AgentDTO{
		ID:           a.ID,
		Name:         a.Name,
		TotalChecked: a.TotalChecked,
		FormID:       a.FormID,
		RegType:      a.RegType,
	}
}
