package application

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/gbaski/gbaski-platform/internal/registration/contracts"
	"github.com/google/uuid"
)

type AgentService struct {
	repo  contracts.RegistrationRepository
	store contracts.CheckInAgentStore
}

func NewAgentService(repo contracts.RegistrationRepository, store contracts.CheckInAgentStore) *AgentService {
	return &AgentService{repo: repo, store: store}
}

func (s *AgentService) GetAgents(ctx context.Context, query GetAgentsQuery) ([]AgentDTO, error) {
	agents, err := s.store.GetAllAgents(ctx, query.FormID)
	if err != nil {
		return nil, err
	}

	if len(agents) == 0 {
		return []AgentDTO{
			{
				Name: "Default Agent",
				ID:   uuid.New(),
			},
		}, nil
	}

	agentList := []AgentDTO{}
	for idStr, name := range agents {
		id, _ := uuid.Parse(idStr)
		agentList = append(agentList, AgentDTO{
			Name: name,
			ID:   id,
		})
	}

	return agentList, nil
}

func (s *AgentService) CreateAgent(ctx context.Context, cmd CreateAgentCommand) ([]AgentDTO, error) {
	id := uuid.New()
	if err := s.store.SetAgent(ctx, cmd.FormID, id, cmd.Name); err != nil {
		return nil, err
	}
	return s.GetAgents(ctx, GetAgentsQuery{FormID: cmd.FormID})
}

func (s *AgentService) DeleteAgent(ctx context.Context, cmd DeleteAgentCommand) ([]AgentDTO, error) {
	id, err := uuid.Parse(cmd.Token)
	if err != nil {
		return nil, err
	}
	if err := s.store.DeleteAgent(ctx, cmd.FormID, id); err != nil {
		return nil, err
	}
	return s.GetAgents(ctx, GetAgentsQuery{FormID: cmd.FormID})
}

func (s *AgentService) GetAgentByToken(ctx context.Context, query GetAgentByTokenQuery) (*AgentDTO, error) {
	decodedToken, err := base64.StdEncoding.DecodeString(query.Token)
	if err != nil {
		return nil, err
	}

	var data struct {
		AgentID uuid.UUID `json:"agentId"`
		FormID  uuid.UUID `json:"formId"`
		RegType string    `json:"regType"`
	}
	if err := json.Unmarshal(decodedToken, &data); err != nil {
		return nil, err
	}

	agentName, err := s.store.GetAgent(ctx, data.FormID, data.AgentID)
	if err != nil {
		return nil, err
	}

	return &AgentDTO{
		Name:    agentName,
		ID:      data.AgentID,
		FormID:  data.FormID,
		RegType: data.RegType,
	}, nil
}
