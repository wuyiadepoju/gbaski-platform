package application

import (
	"time"

	"github.com/gbaski/gbaski-platform/internal/setting/domain"
)

type SettingDTO struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Category  string    `json:"category"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func ToSettingDTO(s domain.Setting) SettingDTO {
	return SettingDTO{
		Key:       s.Key,
		Value:     s.Value,
		Category:  s.Category,
		UpdatedAt: s.UpdatedAt,
	}
}
