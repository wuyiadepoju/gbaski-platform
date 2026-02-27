package form

import (
	"encoding/json"

	"github.com/gbaski/gbaski-ext/auth"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService() *Service {
	return &Service{repo: NewRepository()}
}

type FormHandler struct {
	*Service
}

func NewFormHandler() *FormHandler {
	return &FormHandler{NewService()}
}

func (s *FormHandler) CreateForm(c *fiber.Ctx) error {

	var createFormRequest FormRequest

	if err := c.BodyParser(&createFormRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"status":  "error",
			"error":   err.Error(),
		})
	}

	userId := auth.GetUserId(c)

	formId, err := s.createForm(createFormRequest, userId)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Failed to create form",
			"status":  "error",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Form created successfully",
		"status":  "success",
		"data":    formId,
	})
}

func (s *Service) createForm(form FormRequest, userId uuid.UUID) (*uuid.UUID, error) {

	schema, err := json.Marshal(form.Schema)

	if err != nil {
		return nil, err
	}

	configs, err := json.Marshal(form.Configs)

	if err != nil {
		return nil, err
	}

	// fee, err := json.Marshal(form.Fee)

	// if err != nil {
	// 	return nil, err
	// }

	formModel := FormModel{
		FormRequest: form,
		Schema:      schema,
		Configs:     configs,
		Price:       0.0,
		UserId:      userId,
	}

	if form.Id != nil {

		err = s.repo.UpdateForm(formModel, *form.Id)

		if err != nil {
			return nil, err
		}

		return form.Id, nil
	}

	id, err := s.repo.CreateForm(formModel)

	if err != nil {
		return nil, err
	}

	return id, nil
}

func (s *Service) getFormsByEventId(eventId uuid.UUID) ([]FormItem, error) {
	return s.repo.GetFormsByEventId(eventId)
}
