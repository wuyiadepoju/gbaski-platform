package application

import (
	"context"

	"github.com/gbaski/gbaski-platform/internal/payout/contracts"
)

// GetBanksService fetches the list of available banks
type GetBanksService struct {
	repo contracts.PayoutRepository
}

func NewGetBanksService(repo contracts.PayoutRepository) *GetBanksService {
	return &GetBanksService{repo: repo}
}

// Execute returns the list of banks mapped to DTOs
func (s *GetBanksService) Execute(ctx context.Context) ([]BankDTO, error) {
	banks, err := s.repo.GetBanks(ctx)
	if err != nil {
		return nil, err
	}

	dtos := make([]BankDTO, 0, len(banks))
	for _, b := range banks {
		dtos = append(dtos, ToBankDTO(b))
	}
	return dtos, nil
}
