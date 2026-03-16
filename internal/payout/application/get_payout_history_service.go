package application

import (
	"context"

	"github.com/gbaski/gbaski-platform/internal/payout/contracts"
)

// GetPayoutHistoryService fetches paginated payout history
type GetPayoutHistoryService struct {
	repo contracts.PayoutRepository
}

func NewGetPayoutHistoryService(repo contracts.PayoutRepository) *GetPayoutHistoryService {
	return &GetPayoutHistoryService{repo: repo}
}

// Execute returns a paginated payout history for the given form and user
func (s *GetPayoutHistoryService) Execute(ctx context.Context, query GetPayoutHistoryQuery) (contracts.PayoutList, error) {
	return s.repo.GetPayoutHistory(ctx, query.Params, query.FormID, query.UserID)
}
