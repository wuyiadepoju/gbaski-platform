package domain

import "time"

type RegistrationItem struct {
	ID        *int
	FirstName string
	LastName  *string
	Email     string
	Phone     *string
	RegRef    string
	CreatedAt time.Time
	// ... other fields if needed for domain logic
}

type RegistrationList struct {
	Items []RegistrationItem
	Page  int
	Size  int
	Total int
}

type TicketAttendeeList struct {
	Items []Attendee
	Page  int
	Size  int
	Total int
}

type RegistrationStats struct {
	TotalRegistered int
	TotalCheckedIn  int
	TotalCancelled  int
	TotalAmount     float64
}

type RegistrationQueryModel struct {
	Search string
	Filter map[string]string
	Range  map[string][]string
	Page   int
	Size   int
}
