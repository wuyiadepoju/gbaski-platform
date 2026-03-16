package application

import (
	"context"
	"fmt"

	"github.com/gbaski/gbaski-ext/log"
	"github.com/gbaski/gbaski-platform/internal/payout/contracts"
	"github.com/gbaski/gbaski-platform/internal/payout/domain"
)

// ChangePayoutTypeService changes the payout type on a payout account
type ChangePayoutTypeService struct {
	repo     contracts.PayoutRepository
	eventBus contracts.EventBus
}

func NewChangePayoutTypeService(
	repo contracts.PayoutRepository,
	eventBus contracts.EventBus,
) *ChangePayoutTypeService {
	return &ChangePayoutTypeService{
		repo:     repo,
		eventBus: eventBus,
	}
}

// Execute validates the payout type, persists the change, and returns the updated account DTO
func (s *ChangePayoutTypeService) Execute(ctx context.Context, cmd ChangePayoutTypeCommand) (*PayoutAccountDTO, error) {
	payoutType, err := domain.NewPayoutType(cmd.PayoutType)
	if err != nil {
		return nil, fmt.Errorf("invalid payout type: %w", err)
	}

	account, err := s.repo.ChangePayoutType(ctx, cmd.PayoutAccountID, cmd.UserID, payoutType)
	if err != nil {
		return nil, fmt.Errorf("failed to change payout type: %w", err)
	}

	// Apply domain business method to record the change and raise event
	account.ChangePayoutType(payoutType)
	s.publishDomainEvents(account)

	log.Info("payout", "change_payout_type", map[string]interface{}{
		"account_id":  cmd.PayoutAccountID,
		"user_id":     cmd.UserID,
		"payout_type": payoutType.Value(),
	})

	dto := ToPayoutAccountDTO(account)
	return &dto, nil
}

func (s *ChangePayoutTypeService) publishDomainEvents(account *domain.PayoutAccount) {
	for _, event := range account.DomainEvents() {
		s.eventBus.Publish(event)
	}
	account.ClearDomainEvents()
}
