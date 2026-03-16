package application

import (
	"context"

	"github.com/gbaski/gbaski-platform/internal/setting/contracts"
)

type GetSettingsService struct {
	repo contracts.SettingRepository
}

func NewGetSettingsService(repo contracts.SettingRepository) *GetSettingsService {
	return &GetSettingsService{repo: repo}
}

func (s *GetSettingsService) Execute(ctx context.Context) ([]SettingDTO, error) {
	settings, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	dtos := make([]SettingDTO, len(settings))
	for i, setting := range settings {
		dtos[i] = ToSettingDTO(setting)
	}

	return dtos, nil
}
