package registration

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gbaski/gbaski-event/pkg/event"
	"github.com/gbaski/gbaski-ext/log"
	"github.com/gbaski/gbaski-ext/mailcoach"
	"github.com/google/uuid"
)

type TicketService struct {
	repo      *Repository
	mailcoach *mailcoach.Mailcoach
}

func NewTicketService() *TicketService {
	return &TicketService{NewRepository(), mailcoach.New()}
}

func (s *TicketService) GetRegistrations(queryModel RegistrationQueryModel, formId uuid.UUID, userId uuid.UUID) (RegistrationList, error) {
	return s.repo.GetRegistrations(queryModel, formId, "ticket", userId)
}

func (s *TicketService) GetTicketAttendees(queryModel RegistrationQueryModel, formId uuid.UUID, userId uuid.UUID) (TicketAttendeeList, error) {
	return s.repo.GetTicketAttendees(queryModel, formId, userId)
}

func (s *TicketService) GetRegistrationStats(formId uuid.UUID, userId uuid.UUID) (RegistrationStats, error) {
	return s.repo.GetRegistrationStats(formId, "ticket", userId)
}

func (s *TicketService) GetCheckInAgents(formId uuid.UUID) ([]CheckInAgent, error) {

	key := fmt.Sprintf("checkin_agents:%s", formId)

	agents, err := s.repo.Redis.HGetAll(context.Background(), key).Result()

	if err != nil {
		return nil, err
	}

	if len(agents) == 0 {
		return []CheckInAgent{
			{
				Name:         "Default Agent",
				Id:           uuid.New(),
				TotalChecked: 0,
			},
		}, nil
	}

	agents, err = s.repo.Redis.HGetAll(context.Background(), key).Result()

	if err != nil {
		return nil, err
	}

	agentList := []CheckInAgent{}

	for id, name := range agents {
		agentList = append(agentList, CheckInAgent{
			Name:         name,
			Id:           uuid.MustParse(id),
			TotalChecked: s.repo.CountTicketCheckInsByAgent(formId, name),
		})
	}

	return agentList, nil
}

func (s *TicketService) CreateCheckInAgent(formId uuid.UUID, name string) ([]CheckInAgent, error) {

	key := fmt.Sprintf("checkin_agents:%s", formId)

	id := uuid.New().String()

	s.repo.Redis.HSet(context.Background(), key, id, name)
	s.repo.Redis.HExpire(context.Background(), key, 1*time.Hour)

	return s.GetCheckInAgents(formId)
}

func (s *TicketService) DeleteCheckInAgent(formId uuid.UUID, token string) ([]CheckInAgent, error) {

	key := fmt.Sprintf("checkin_agents:%s", formId)

	s.repo.Redis.HDel(context.Background(), key, token)

	return s.GetCheckInAgents(formId)
}

func (s *TicketService) GetCheckInAgent(token string) (*CheckInAgent, error) {

	var data struct {
		AgentId uuid.UUID `json:"agentId"`
		FormId  uuid.UUID `json:"formId"`
		RegType string    `json:"regType"`
	}

	decodedToken, err := base64.StdEncoding.DecodeString(token)

	if err != nil {
		return nil, err
	}

	json.Unmarshal(decodedToken, &data)

	key := fmt.Sprintf("checkin_agents:%s", data.FormId)

	agentName, err := s.repo.Redis.HGet(context.Background(), key, data.AgentId.String()).Result()

	if err != nil {

		if err.Error() == "redis: nil" {
			return nil, errors.New("agent_not_found")
		}

		return nil, err
	}

	totalChecked := s.repo.CountTicketCheckInsByAgent(data.FormId, agentName)

	return &CheckInAgent{
		Name:         agentName,
		Id:           data.AgentId,
		TotalChecked: totalChecked,
		FormId:       data.FormId,
		RegType:      data.RegType,
	}, nil
}

func (s *TicketService) LookupRegistration(formId uuid.UUID, regRef string) (*TicketAttendee, error) {

	return s.repo.LookupTicketRegistration(formId, regRef)
}

func (s *TicketService) CheckInTicketRegistration(regRef string, agentName string) error {

	attendee, err := s.repo.CheckInTicketRegistration(regRef, agentName)

	if err != nil {
		return err
	}

	go func() {

		// Use the event service to send check-in email
		eventService := event.NewEventScheduleService()

		emailData := event.TicketCheckInEmailData{
			FirstName:  attendee.FirstName,
			LastName:   attendee.LastName,
			Email:      attendee.Email,
			TicketName: attendee.TicketName,
			RegRef:     attendee.RegRef,
			CheckedAt:  attendee.CheckedAt,
		}

		err := eventService.SendTicketCheckInEmail(emailData)
		if err != nil {
			log.Error("registration", "check_in_ticket_registration", err)
		}

	}()

	return nil
}
