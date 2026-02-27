package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gbaski/gbaski-event/pkg/event"
	"github.com/gbaski/gbaski-event/pkg/thirdparty/google"
	"github.com/gbaski/gbaski-platform/app/form"

	"github.com/google/uuid"

	"github.com/gbaski/gbaski-ext/auth"
	"github.com/gbaski/gbaski-ext/job"
	"github.com/gbaski/gbaski-ext/log"
	"github.com/gbaski/gbaski-ext/mailcoach"
	"github.com/gbaski/gbaski-shared/fee"
	"github.com/gbaski/gbaski-shared/util"
)

type Service struct {
	repo        *Repository
	formHandler *form.FormHandler
	jobService  *job.Service
	mailcoach   *mailcoach.Mailcoach
	authHandler *auth.AuthHandler
}

func NewService() *Service {
	return &Service{repo: NewRepository()}
}

func (s *Service) createUpdateEvent(event EventModel, user User) (*CreateEventResponse, error) {

	status := EventStatusDraft

	configs := json.RawMessage("{}")

	var schema json.RawMessage

	// if event.RegType == RegistrationTypeEntry {
	// 	schema, _ = json.Marshal([]FormSectionContent{})
	// }

	if event.RegType == RegistrationTypeTicket && event.Payment == EventPaymentFree {

		description := "Standard ticket at regular price"
		id := uuid.New()

		limit := 1

		tickets := []SchemaTicketItem{
			{
				Id:          0,
				UUID:        util.Ptr(id),
				Name:        "Regular",
				Description: &description,
				Payment:     TicketPaymentFree,
				Price:       0,
				Discount:    nil,
				Capacity:    nil,
				Limit:       &limit,
				Sold:        0,
				Index:       0,
				Deleted:     false,
			},
		}

		schema, _ = json.Marshal(tickets)
	}

	event.AccessType = AccessTypePublic

	if event.AccessType == AccessTypePublic && event.Payment == EventPaymentFree {
		status = EventStatusPublished
	}

	event.Slug = util.Slugify(event.Name)
	event.Status = status
	event.Configs = configs
	event.Schema = schema
	event.UserId = user.Id

	if event.Location == nil || *event.Location == "" {

		if event.ModeType == EventModePhysical {
			event.Location = util.Ptr("Undisclosed Location")
		} else {
			event.Location = util.Ptr("http://virtual")
		}

	}

	otherCategory := event.OtherCategory

	if otherCategory != nil && *otherCategory != "" {

		categoryId, err := s.repo.CreateEventCategory(EventCategory{
			Name:      *otherCategory,
			EventType: "Other",
			Public:    false,
		})

		if err != nil {
			return nil, err
		}

		event.CategoryId = &categoryId
	}

	if event.SocialMedia == nil || string(event.SocialMedia) == "" {
		event.SocialMedia = json.RawMessage("{}")
	}

	if event.Id == nil {

		eventItem, err := s.createEvent(event)

		if err != nil {
			return nil, err
		}

		event.FormId = &eventItem.FormId
		event.Id = eventItem.Id

		go s.scheduleEventEmail(event)

		log.Info("event", "create_event", map[string]any{
			"form_id": eventItem.FormId,
		})
	} else {

		_event, err := s.repo.GetEvent(*event.Id)

		if err != nil {
			return nil, err
		}

		event.Payment = _event.Payment

		err = s.updateEvent(event)

		if err != nil {
			return nil, err
		}

		go s.reScheduleEventEmail(_event, event.StartDate, event.EndDate)
	}

	formId := event.FormId

	return &CreateEventResponse{
		FormId:       *formId,
		EventUrl:     util.EventUrl(user.Name, event.Name),
		EventStatus:  status,
		EventPayment: event.Payment,
	}, nil
}

func (s *Service) createEvent(event EventModel) (*EventItem, error) {

	go func() {

		webmasterService, err := google.NewWebmasterService(context.Background())
		if err != nil {
			log.Error("webmaster", "init_webmaster_service", err)
			return
		}

		eventUrl := util.EventUrl(event.BrandName, event.Name)

		err = webmasterService.Add(eventUrl)
		if err != nil {
			log.Error("webmaster", "add_event_url", err)
			return
		}

	}()

	return s.repo.CreateEvent(event)
}

func (s *Service) updateEvent(event EventModel) error {

	event.Slug = util.Slugify(event.Name)
	return s.repo.UpdateEvent(event)
}

func (s *Service) getEventStats(eventId uuid.UUID, userId uuid.UUID) ([]EventStatsItem, error) {

	stats, err := s.repo.GetEventStats(eventId, userId)

	if err != nil {
		return nil, err
	}

	return stats, nil

}

func (s *Service) getEvent(id uuid.UUID) (*EventItem, error) {

	event, err := s.repo.GetEvent(id)

	if err != nil {
		return nil, err
	}

	// features, err := s.GetEventFeatures(id)

	// if err != nil {
	// 	return nil, err
	// }

	// return &EventDetail{
	// 	Item:     *event,
	// 	Features: features,
	// }, nil

	return event, nil
}

func (s *Service) getEvents(queryModel EventQueryModel, userId uuid.UUID) (EventList, error) {
	return s.repo.GetEvents(queryModel, userId)
}

func (s *Service) GetEventFeatures(eventId uuid.UUID) ([]EventFeature, error) {

	var features []EventFeature

	features, err := s.repo.GetEventFeatures(eventId)

	if err != nil {
		return nil, err
	}

	return features, nil
}

func (s *Service) updateEventStatus(eventId uuid.UUID, status EventStatus) (*EventItem, error) {
	return s.repo.UpdateEventStatus(eventId, status)
}

func (s *Service) updateEventImage(eventId uuid.UUID, imageURL string) error {
	go InvalidateStorage(imageURL)
	return s.repo.UpdateEventImage(eventId, imageURL)
}

func (s *Service) updateEventVideo(eventId uuid.UUID, videoURL string) error {
	go InvalidateStorage(videoURL)
	return s.repo.UpdateEventVideo(eventId, videoURL)
}

func (s *Service) deleteEvent(eventId uuid.UUID) error {
	return s.repo.DeleteEvent(eventId)
}

func (s *Service) getEventCategories() ([]EventCategory, error) {
	return s.repo.GetEventCategories()
}

func (s *Service) getEventForms(userId uuid.UUID) ([]EventForm, error) {

	forms, err := s.repo.GetEventForms(userId)

	if err != nil {
		return nil, err
	}

	for index, form := range forms {

		if form.DirectPaymentEnabled {

			payoutProvider := s.repo.GetPayoutProviderByUserId(form.UserId)

			if payoutProvider != nil {
				forms[index].Payout = &event.PayoutProviderItem{
					Provider:  payoutProvider.Provider,
					PublicKey: payoutProvider.PublicKey,
				}

				forms[index].FeeRate = "0.00"
				forms[index].FeePayer = FeePayerWallet
			}
		} else {
			if forms[index].FeePayer != FeePayerBuyer && forms[index].FeePayer != FeePayerSeller {
				forms[index].FeePayer = FeePayerSeller
			}
		}

		forms[index].EventUrl = util.EventUrl(form.BrandName, form.EventName)
	}

	return forms, nil
}

func (s *Service) createForm(form FormRequest, user User) (*uuid.UUID, error) {

	configs, err := json.Marshal(form.Configs)

	if err != nil {
		return nil, err
	}

	directPaymentEnabled := s.repo.DirectPaymentEnabled(user.Id)

	if form.FeePayer == FeePayerBuyer && !directPaymentEnabled {
		saleAmount, _ := fee.BuyerTicketFeeInsert(form.Price)
		form.Price = saleAmount
	}

	formModel := FormModel{
		FormRequest: form,
		Price:       form.Price,
		Configs:     configs,
		UserId:      user.Id,
	}

	formId, err := s.repo.CreateEventForm(formModel)

	if err != nil {
		return nil, err
	}

	event, err := s.updateEventStatus(form.EventId, EventStatusPublished)

	if err != nil {
		return nil, err
	}

	eventModel := EventModel{}

	eventModel.Id = event.Id
	eventModel.UserId = user.Id
	eventModel.Status = EventStatusPublished
	eventModel.RegType = form.RegType
	eventModel.FormId = formId
	eventModel.Name = event.Name
	eventModel.Description = event.Description
	eventModel.Slug = util.Slugify(event.Name)
	eventModel.Configs = configs
	eventModel.Payment = event.Payment
	eventModel.ModeType = event.ModeType
	eventModel.DurationType = event.DurationType
	eventModel.StartDate = event.StartDate
	eventModel.EndDate = event.EndDate
	eventModel.Location = event.Location

	go s.scheduleEventEmail(eventModel)

	return formId, nil
}

func (s *Service) getFormsByEventId(eventId uuid.UUID) ([]FormItem, error) {
	return s.repo.GetFormsByEventId(eventId)
}

func (s *Service) getEventsReport(userId uuid.UUID) (*EventsReport, error) {
	return s.repo.GetEventsReport(userId)
}

func (s *Service) scheduleEventEmail(event EventModel) {

	type ScheduleItem struct {
		Date        time.Time
		Name        string
		Description string
		FuncName    string
		FuncParam   map[string]any
	}

	// Calculate dates based on event start and end dates
	eventStart := event.StartDate
	eventEnd := event.EndDate
	now := time.Now()

	// Get event creation date - use CreatedAt if event exists, otherwise use now
	var eventCreatedAt time.Time
	if event.Id != nil {
		existingEvent, err := s.repo.GetEvent(*event.Id)
		if err == nil && existingEvent != nil {
			eventCreatedAt = existingEvent.CreatedAt
		} else {
			// If event doesn't exist yet or can't be fetched, use current time
			eventCreatedAt = now
		}
	} else {
		// Event is being created now
		eventCreatedAt = now
	}

	user, err := s.authHandler.GetUserById(event.UserId)
	if err != nil {
		log.Error("event", "get_user_by_id", err)
		return
	}

	organizerSchedules := []ScheduleItem{
		{
			Date:        eventCreatedAt.Add(1 * time.Minute), // Send 1 minute after event creation
			Name:        "event-created",
			Description: "Create email list, send transactional email, and welcome email when event is created",
			FuncName:    "SendEventCreated",
			FuncParam: map[string]any{
				"eventId": event.Id,
			},
		},
	}

	organizerSchedules = append(organizerSchedules, ScheduleItem{
		Date:        eventCreatedAt.Add(3 * 24 * time.Hour),
		Name:        "host-event-engagement-tips",
		Description: "3 days after event creation - event engagement tips and tasks",
		FuncName:    "SendEventEngagementTips",
		FuncParam: map[string]any{
			"eventId": event.Id,
			"email":   user.Email,
		},
	})

	organizerSchedules = append(organizerSchedules, ScheduleItem{
		Date:        eventStart.Add(-7 * 24 * time.Hour),
		Name:        "host-pre-event-reminder",
		Description: "1 week before event - pre-event engagement reminder",
		FuncName:    "RemindOrganizerPreEvent",
		FuncParam: map[string]any{
			"eventId": event.Id,
			"email":   user.Email,
		},
	})

	organizerSchedules = append(organizerSchedules, ScheduleItem{
		Date:        eventEnd.Add(10 * time.Minute),
		Name:        "host-update-event-status-ended",
		Description: "10 minutes after event end - update event status to ended",
		FuncName:    "UpdateEventStatusToEnded",
		FuncParam: map[string]any{
			"eventId": event.Id,
		},
	})

	organizerSchedules = append(organizerSchedules, ScheduleItem{
		Date:        eventEnd.Add(48 * time.Hour),
		Name:        "host-post-event-recap",
		Description: "48 hours after event - post-event recap and insights",
		FuncName:    "SendPostEventRecap",
		FuncParam: map[string]any{
			"eventId": event.Id,
			"email":   user.Email,
		},
	})

	// Participant email schedules
	participantSchedules := []ScheduleItem{
		{
			Date:        eventStart.Add(-5 * 24 * time.Hour),
			Name:        "participant-pre-event-reminder",
			Description: "5 days before event - early reminder",
			FuncName:    "RemindParticipantPreEvent",
			FuncParam: map[string]any{
				"eventId": event.Id,
			},
		},
		{
			Date:        eventStart.Add(-3 * 24 * time.Hour),
			Name:        "participant-three-days-reminder",
			Description: "3 days before event - countdown reminder",
			FuncName:    "RemindParticipantPreEvent",
			FuncParam: map[string]any{
				"eventId": event.Id,
			},
		},
		{
			Date:        eventStart.Add(-24 * time.Hour),
			Name:        "participant-one-day-reminder",
			Description: "1 day before event - final reminder",
			FuncName:    "RemindParticipantPreEvent",
			FuncParam: map[string]any{
				"eventId": event.Id,
			},
		},
		{
			Date:        eventStart.Add(-3 * time.Hour),
			Name:        "participant-event-day-reminder",
			Description: "3 hours before event",
			FuncName:    "RemindParticipantPreEvent",
			FuncParam: map[string]any{
				"eventId": event.Id,
			},
		},
		{
			Date:        eventEnd.Add(2 * time.Hour),
			Name:        "participant-post-event",
			Description: "2 hours after event - thank you and feedback",
			FuncName:    "ThankParticipantPostEvent",
			FuncParam: map[string]any{
				"eventId": event.Id,
			},
		},
	}

	// Combine all schedules
	allSchedules := append(organizerSchedules, participantSchedules...)

	// Delete existing scheduled jobs for this event to prevent duplicates
	eventJobPrefix := fmt.Sprintf("event-email-%s-", event.Id.String())
	err = s.jobService.DeleteJobsByNamePrefix(eventJobPrefix)
	if err != nil {
		log.Error("event", "delete_existing_scheduled_jobs", err)
	} else {
		log.Info("event", "deleted_existing_scheduled_jobs", map[string]interface{}{
			"event_id":   event.Id,
			"job_prefix": eventJobPrefix,
		})
	}

	// Filter out schedules that are in the past
	var validSchedules []ScheduleItem
	for _, scheduleItem := range allSchedules {
		if scheduleItem.Date.After(now) {
			validSchedules = append(validSchedules, scheduleItem)
		} else {
			log.Info("event", "schedule_filtered_past", map[string]interface{}{
				"event_id":       event.Id,
				"schedule":       scheduleItem.Name,
				"scheduled_date": scheduleItem.Date,
				"current_date":   now,
			})
		}
	}

	// Create jobs for valid schedules
	for _, scheduleItem := range validSchedules {

		jobName := fmt.Sprintf("event-email-%s-%s", event.Id.String(), scheduleItem.Name)
		jobDescription := fmt.Sprintf("%s - %s", event.Name, scheduleItem.Description)

		err := s.jobService.CreateJob(&job.JobItem{
			Date:        scheduleItem.Date,
			Name:        jobName,
			Description: &jobDescription,
			Type:        job.JobTypeEvent,
			Func: job.JobFunc{
				FuncName:  scheduleItem.FuncName,
				FuncParam: scheduleItem.FuncParam,
			},
			Status:   job.JobStatusScheduled,
			TimeZone: "UTC",
		})

		if err != nil {
			log.Error("event", "create_scheduled_job", err)
		} else {
			log.Info("event", "schedule_event_email", map[string]interface{}{
				"event_id":    event.Id,
				"job_name":    jobName,
				"date":        scheduleItem.Date,
				"funcName":    scheduleItem.FuncName,
				"description": scheduleItem.Description,
			})
		}
	}

	log.Info("event", "schedule_event_email_complete", map[string]interface{}{
		"event_id":         event.Id,
		"total_schedules":  len(allSchedules),
		"valid_schedules":  len(validSchedules),
		"event_start_date": eventStart,
		"event_end_date":   eventEnd,
	})
}

func (s *Service) reScheduleEventEmail(currentEvent *EventItem, newStartDate, newEndDate time.Time) {

	// Check if dates have actually changed
	if currentEvent.StartDate.Equal(newStartDate) && currentEvent.EndDate.Equal(newEndDate) {
		log.Info("event", "schedule_update_skipped", map[string]interface{}{
			"event_id": currentEvent.Id,
			"reason":   "dates_unchanged",
		})
		return
	}

	log.Info("event", "schedule_update_started", map[string]interface{}{
		"event_id":       currentEvent.Id,
		"old_start_date": currentEvent.StartDate,
		"new_start_date": newStartDate,
		"old_end_date":   currentEvent.EndDate,
		"new_end_date":   newEndDate,
	})

	// Delete existing scheduled jobs for this event
	eventJobPrefix := fmt.Sprintf("event-email-%s-", currentEvent.Id.String())
	err := s.jobService.DeleteJobsByNamePrefix(eventJobPrefix)
	if err != nil {
		log.Error("event", "delete_existing_scheduled_jobs", err)
	} else {
		log.Info("event", "deleted_existing_scheduled_jobs", map[string]interface{}{
			"event_id":   currentEvent.Id,
			"job_prefix": eventJobPrefix,
		})
	}

	// Create new event model with updated dates for scheduling
	updatedEvent := EventModel{
		Id:     currentEvent.Id,
		UserId: currentEvent.UserId,
		CreateEventRequest: CreateEventRequest{
			Name:      currentEvent.Name,
			StartDate: newStartDate,
			EndDate:   newEndDate,
		},
	}

	// Reschedule all email jobs with new dates
	s.scheduleEventEmail(updatedEvent)

	log.Info("event", "schedule_update_completed", map[string]interface{}{
		"event_id":       currentEvent.Id,
		"new_start_date": newStartDate,
		"new_end_date":   newEndDate,
	})
}
