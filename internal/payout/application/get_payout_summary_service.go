package application

import (
	"context"

	"github.com/gbaski/gbaski-platform/internal/payout/contracts"
)

// GetPayoutSummaryService fetches the payout summary for a form
type GetPayoutSummaryService struct {
	repo contracts.PayoutRepository
}

func NewGetPayoutSummaryService(repo contracts.PayoutRepository) *GetPayoutSummaryService {
	return &GetPayoutSummaryService{repo: repo}
}

// Execute returns the payout summary grouped by payout type
func (s *GetPayoutSummaryService) Execute(ctx context.Context, query GetPayoutSummaryQuery) (map[string]contracts.PayoutSummaryDTO, error) {
	return s.repo.GetPayoutSummary(ctx, query.FormID, query.UserID)
}
