package repo

import (
	"context"
	"database/sql"

	"github.com/gbaski/gbaski-platform/internal/setting/contracts"
	"github.com/gbaski/gbaski-platform/internal/setting/domain"
	"github.com/gbaski/gbaski-shared/repo"
)

type SettingRepositoryImpl struct {
	*repo.BaseRepository
}

func NewSettingRepositoryImpl() contracts.SettingRepository {
	return &SettingRepositoryImpl{repo.NewBaseRepository()}
}

func (r *SettingRepositoryImpl) GetByKey(ctx context.Context, key string) (*domain.Setting, error) {
	var s domain.Setting
	query := "SELECT key, value, category, created_at, updated_at FROM settings WHERE key = $1"
	err := r.DB.GetContext(ctx, &s, query, key)
	if err == sql.ErrNoRows {
		return nil, domain.ErrSettingNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SettingRepositoryImpl) GetAll(ctx context.Context) ([]domain.Setting, error) {
	var settings []domain.Setting
	query := "SELECT key, value, category, created_at, updated_at FROM settings"
	err := r.DB.SelectContext(ctx, &settings, query)
	if err != nil {
		return nil, err
	}
	return settings, nil
}

func (r *SettingRepositoryImpl) Save(ctx context.Context, s *domain.Setting) error {
	query := `
		INSERT INTO settings (key, value, category, created_at, updated_at)
		VALUES (:key, :value, :category, :created_at, :updated_at)
		ON CONFLICT (key) DO UPDATE SET
			value = EXCLUDED.value,
			category = EXCLUDED.category,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.DB.NamedExecContext(ctx, query, s)
	return err
}
