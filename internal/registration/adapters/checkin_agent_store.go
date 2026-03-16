package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/gbaski/gbaski-platform/internal/registration/contracts"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type CheckInAgentStoreImpl struct {
	redis *redis.Client
}

func NewCheckInAgentStoreImpl(redis *redis.Client) contracts.CheckInAgentStore {
	return &CheckInAgentStoreImpl{redis: redis}
}

func (s *CheckInAgentStoreImpl) SetAgent(ctx context.Context, formID uuid.UUID, id uuid.UUID, name string) error {
	key := fmt.Sprintf("checkin_agents:%s", formID)
	if err := s.redis.HSet(ctx, key, id.String(), name).Err(); err != nil {
		return err
	}
	return s.redis.Expire(ctx, key, 1*time.Hour).Err()
}

func (s *CheckInAgentStoreImpl) GetAgent(ctx context.Context, formID uuid.UUID, id uuid.UUID) (string, error) {
	key := fmt.Sprintf("checkin_agents:%s", formID)
	name, err := s.redis.HGet(ctx, key, id.String()).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("agent_not_found")
	}
	return name, err
}

func (s *CheckInAgentStoreImpl) DeleteAgent(ctx context.Context, formID uuid.UUID, id uuid.UUID) error {
	key := fmt.Sprintf("checkin_agents:%s", formID)
	return s.redis.HDel(ctx, key, id.String()).Err()
}

func (s *CheckInAgentStoreImpl) GetAllAgents(ctx context.Context, formID uuid.UUID) (map[string]string, error) {
	key := fmt.Sprintf("checkin_agents:%s", formID)
	return s.redis.HGetAll(ctx, key).Result()
}
