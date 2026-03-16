package setting

import (
	"github.com/gbaski/gbaski-platform/internal/common"
	"github.com/gbaski/gbaski-platform/internal/setting/application"
	"github.com/gbaski/gbaski-platform/internal/setting/repo"
	"github.com/gofiber/fiber/v2"
)

type Response = common.Response

type SettingHandler struct {
	getSettingsService   *application.GetSettingsService
	updateSettingService *application.UpdateSettingService
}

func NewSettingHandler() *SettingHandler {
	repository := repo.NewSettingRepositoryImpl()
	return &SettingHandler{
		getSettingsService:   application.NewGetSettingsService(repository),
		updateSettingService: application.NewUpdateSettingService(repository),
	}
}

func (h *SettingHandler) GetSettings(c *fiber.Ctx) error {
	settings, err := h.getSettingsService.Execute(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: "Failed to fetch settings",
			Error:   err.Error(),
		})
	}

	return c.JSON(Response{
		Status:  "success",
		Message: "Settings fetched successfully",
		Data:    settings,
	})
}

func (h *SettingHandler) UpdateSetting(c *fiber.Ctx) error {
	var cmd application.UpdateSettingCommand
	if err := c.BodyParser(&cmd); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	if err := h.updateSettingService.Execute(c.Context(), cmd); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Status:  "error",
			Message: "Failed to update setting",
			Error:   err.Error(),
		})
	}

	return c.JSON(Response{
		Status:  "success",
		Message: "Setting updated successfully",
	})
}
