package contracts

import (
	"context"

	"github.com/gbaski/gbaski-platform/internal/setting/domain"
)

type SettingRepository interface {
	GetByKey(ctx context.Context, key string) (*domain.Setting, error)
	GetAll(ctx context.Context) ([]domain.Setting, error)
	Save(ctx context.Context, setting *domain.Setting) error
}
