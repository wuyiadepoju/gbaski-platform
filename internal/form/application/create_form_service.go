package application

import (
	"context"
	"encoding/json"

	"github.com/gbaski/gbaski-ext/log"
	"github.com/gbaski/gbaski-platform/internal/form/contracts"
	"github.com/gbaski/gbaski-platform/internal/form/domain"
	eventdomain "github.com/gbaski/gbaski-platform/internal/event/domain"
	"github.com/google/uuid"
)

const formComponent = "form"

// CreateFormCommand is the command used by the application service
type CreateFormCommand struct {
	UserID uuid.UUID
	CreateFormRequest
}

// CreateFormService handles creation and update of forms
type CreateFormService struct {
	repo     contracts.FormRepository
	eventBus contracts.EventBus
}

func NewCreateFormService(repo contracts.FormRepository, eventBus contracts.EventBus) *CreateFormService {
	return &CreateFormService{
		repo:     repo,
		eventBus: eventBus,
	}
}

// Execute creates or updates a form based on presence of ID
func (s *CreateFormService) Execute(ctx context.Context, cmd CreateFormCommand) (*uuid.UUID, error) {
	schema, err := json.Marshal(cmd.Schema)
	if err != nil {
		log.Error(formComponent, "marshal_schema", err)
		return nil, err
	}

	form := &domain.FormAggregate{
		ID:            cmd.Id,
		EventID:       cmd.EventId,
		Name:          cmd.Name,
		RegType:       cmd.RegType,
		IsPrimaryForm: cmd.IsPrimaryForm,
		Price:         cmd.Price,
		Currency:      cmd.Currency,
		BuyerPaysFee:  cmd.BuyerPaysFee,
		AccessType:    cmd.AccessType,
		Schema:        schema,
		Configs:       cmd.Configs,
		StartDate:     cmd.StartDate,
		EndDate:       cmd.EndDate,
		UserID:        cmd.UserID,
	}

	if form.ID != nil {
		if err := s.repo.UpdateForm(ctx, form); err != nil {
			log.Error(formComponent, "update_form", err)
			return nil, err
		}

		// Publish domain event for form update (reusing generic DomainEvent)
		s.eventBus.Publish(eventdomain.NewEventUpdatedEvent(*form.ID, form.UserID))

		log.Info(formComponent, "update_form", map[string]interface{}{
			"form_id":  form.ID,
			"event_id": form.EventID,
			"user_id":  form.UserID,
		})
		return form.ID, nil
	}

	id, err := s.repo.CreateForm(ctx, form)
	if err != nil {
		log.Error(formComponent, "create_form", err)
		return nil, err
	}

	// Publish domain event for form creation (reusing generic DomainEvent)
	if id != nil {
		s.eventBus.Publish(eventdomain.NewEventCreatedEvent(*id, form.UserID, form.Name))
	}

	log.Info(formComponent, "create_form", map[string]interface{}{
		"form_id":  id,
		"event_id": form.EventID,
		"user_id":  form.UserID,
	})

	return id, nil
}
