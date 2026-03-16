package application

import (
	"context"

	"github.com/gbaski/gbaski-platform/internal/payout/contracts"
)

// GetTotalPayoutService fetches the total payout amount for a form
type GetTotalPayoutService struct {
	repo contracts.PayoutRepository
}

func NewGetTotalPayoutService(repo contracts.PayoutRepository) *GetTotalPayoutService {
	return &GetTotalPayoutService{repo: repo}
}

// Execute returns the total payout amount for a given user and form
func (s *GetTotalPayoutService) Execute(ctx context.Context, query GetTotalPayoutQuery) float64 {
	return s.repo.GetTotalPayout(ctx, query.UserID, query.FormID)
}
