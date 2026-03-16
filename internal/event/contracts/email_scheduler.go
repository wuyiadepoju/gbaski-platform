package contracts

import (
	"time"
	"github.com/gbaski/gbaski-platform/internal/event/domain"
)

// EmailScheduler defines the contract for scheduling event-related emails
type EmailScheduler interface {
	ScheduleEventEmails(event *domain.Event)
	RescheduleEventEmails(event *domain.Event, newStartDate, newEndDate time.Time)
}
