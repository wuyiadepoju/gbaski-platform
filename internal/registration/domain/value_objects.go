package domain

type RegistrationStatus string

const (
	RegistrationStatusRegistered RegistrationStatus = "registered"
	RegistrationStatusCheckedIn  RegistrationStatus = "checkedin"
	RegistrationStatusCancelled  RegistrationStatus = "cancelled"
)

func NewRegistrationStatus(status string) (RegistrationStatus, error) {
	switch RegistrationStatus(status) {
	case RegistrationStatusRegistered, RegistrationStatusCheckedIn, RegistrationStatusCancelled:
		return RegistrationStatus(status), nil
	default:
		return "", ErrInvalidRegStatus
	}
}

func (r RegistrationStatus) String() string {
	return string(r)
}
