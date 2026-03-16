package application

import (
	"context"

	"github.com/gbaski/gbaski-platform/internal/payout/contracts"
)

// GetPayoutAccountService fetches the payout account for a user
type GetPayoutAccountService struct {
	repo contracts.PayoutRepository
}

func NewGetPayoutAccountService(repo contracts.PayoutRepository) *GetPayoutAccountService {
	return &GetPayoutAccountService{repo: repo}
}

// Execute returns the payout account DTO for the given currency and user
func (s *GetPayoutAccountService) Execute(ctx context.Context, query GetPayoutAccountQuery) (*PayoutAccountDTO, error) {
	account, err := s.repo.FindPayoutAccount(ctx, query.Currency, query.UserID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, nil
	}
	dto := ToPayoutAccountDTO(account)
	return &dto, nil
}
