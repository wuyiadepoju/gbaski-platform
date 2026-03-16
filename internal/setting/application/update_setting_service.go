package application

import (
	"context"

	"github.com/gbaski/gbaski-platform/internal/setting/contracts"
	"github.com/gbaski/gbaski-platform/internal/setting/domain"
)

type UpdateSettingService struct {
	repo contracts.SettingRepository
}

func NewUpdateSettingService(repo contracts.SettingRepository) *UpdateSettingService {
	return &UpdateSettingService{repo: repo}
}

func (s *UpdateSettingService) Execute(ctx context.Context, cmd UpdateSettingCommand) error {
	setting, err := s.repo.GetByKey(ctx, cmd.Key)
	if err != nil && err != domain.ErrSettingNotFound {
		return err
	}

	if setting == nil {
		// Create new setting if it doesn't exist
		setting, err = domain.NewSetting(cmd.Key, cmd.Value, "general")
		if err != nil {
			return err
		}
	} else {
		setting.UpdateValue(cmd.Value)
	}

	return s.repo.Save(ctx, setting)
}
