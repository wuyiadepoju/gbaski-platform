package application

import (
	"context"

	"github.com/gbaski/gbaski-platform/internal/registration/contracts"
	"github.com/gbaski/gbaski-platform/internal/registration/domain"
)

type GetRegistrationsService struct {
	repo contracts.RegistrationRepository
}

func NewGetRegistrationsService(repo contracts.RegistrationRepository) *GetRegistrationsService {
	return &GetRegistrationsService{repo: repo}
}

func (s *GetRegistrationsService) Execute(ctx context.Context, query GetRegistrationsQuery) (domain.RegistrationList, error) {
	// The repo returns domain.RegistrationList directly for now to preserve pager logic
	return s.repo.GetRegistrations(ctx, query.Query, query.FormID, "ticket", query.UserID)
}

type GetRegistrationStatsService struct {
	repo contracts.RegistrationRepository
}

func NewGetRegistrationStatsService(repo contracts.RegistrationRepository) *GetRegistrationStatsService {
	return &GetRegistrationStatsService{repo: repo}
}

func (s *GetRegistrationStatsService) Execute(ctx context.Context, query GetRegistrationStatsQuery) (domain.RegistrationStats, error) {
	return s.repo.GetRegistrationStats(ctx, query.FormID, "ticket", query.UserID)
}
