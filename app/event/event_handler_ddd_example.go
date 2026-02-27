package event

import (
	"fmt"
	"strings"

	"github.com/gbaski/gbaski-platform/app/event/application"
	"github.com/gbaski/gbaski-platform/app/event/infrastructure"
	"github.com/gbaski/gbaski-ext/auth"
	"github.com/gbaski/gbaski-ext/job"
	"github.com/gbaski/gbaski-host/internal/common"
	"github.com/gbaski/gbaski-shared/pager"
	"github.com/gbaski/gbaski-shared/validator"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// EventHandlerDDD is a DDD-compliant event handler example
// This shows how to wire application services and use them in handlers
type EventHandlerDDD struct {
	createService        *application.CreateEventService
	getService           *application.GetEventService
	getEventsService     *application.GetEventsService
	updateService        *application.UpdateEventService
	updateStatusService  *application.UpdateEventStatusService
	deleteService        *application.DeleteEventService
	updateImageService   *application.UpdateEventImageService
	updateVideoService   *application.UpdateEventVideoService
}

// NewEventHandlerDDD creates a new DDD-compliant event handler
// This demonstrates dependency injection and wiring
func NewEventHandlerDDD() *EventHandlerDDD {
	// Initialize infrastructure
	eventRepo := infrastructure.NewEventRepositoryImpl()
	categoryRepo := infrastructure.NewEventCategoryRepositoryImpl()
	eventBus := infrastructure.NewEventBusImpl()
	webmasterService := infrastructure.NewWebmasterServiceImpl()
	emailScheduler := infrastructure.NewEmailSchedulerImpl(
		job.NewService(),
		auth.NewAuthHandler(),
	)

	// Initialize application services
	createService := application.NewCreateEventService(
		eventRepo,
		categoryRepo,
		eventBus,
		webmasterService,
		emailScheduler,
	)

	getService := application.NewGetEventService(eventRepo)
	getEventsService := application.NewGetEventsService(eventRepo)

	updateService := application.NewUpdateEventService(
		eventRepo,
		categoryRepo,
		eventBus,
		emailScheduler,
	)

	updateStatusService := application.NewUpdateEventStatusService(
		eventRepo,
		eventBus,
	)

	deleteService := application.NewDeleteEventService(
		eventRepo,
		eventBus,
	)

	updateImageService := application.NewUpdateEventImageService(
		eventRepo,
		eventBus,
	)

	updateVideoService := application.NewUpdateEventVideoService(
		eventRepo,
		eventBus,
	)

	return &EventHandlerDDD{
		createService:       createService,
		getService:          getService,
		getEventsService:    getEventsService,
		updateService:       updateService,
		updateStatusService: updateStatusService,
		deleteService:       deleteService,
		updateImageService:  updateImageService,
		updateVideoService:  updateVideoService,
	}
}

// CreateEventDDD demonstrates DDD-compliant event creation
func (h *EventHandlerDDD) CreateEventDDD(c *fiber.Ctx) error {
	var eventRequest CreateEventRequest

	if err := c.BodyParser(&eventRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.Response{
			Status:  "error",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	// Validate request
	isValidData := validator.Validate(eventRequest)
	if !isValidData {
		return c.Status(fiber.StatusBadRequest).JSON(common.Response{
			Status:  "error",
			Message: validator.ErrorMessage(),
		})
	}

	user := auth.GetUser(c)

	// Map request to command
	cmd := application.CreateEventCommand{
		UserID:          user.Id,
		Name:            eventRequest.Name,
		Description:     eventRequest.Description,
		StartDate:       eventRequest.StartDate,
		EndDate:         eventRequest.EndDate,
		Location:        eventRequest.Location,
		ModeType:        string(eventRequest.ModeType),
		Payment:         string(eventRequest.Payment),
		AccessType:      string(eventRequest.AccessType),
		DurationType:    string(eventRequest.DurationType),
		RegistrationType: string(eventRequest.RegType),
		CategoryId:      eventRequest.CategoryId,
		OtherCategory:   eventRequest.OtherCategory,
		BrandName:       user.Name,
	}

	// Execute use case
	result, err := h.createService.Execute(c.Context(), cmd)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "duplicate") && strings.Contains(errStr, "events_name_user_id_unique") {
			return c.Status(fiber.StatusConflict).JSON(common.Response{
				Status:  "error",
				Message: fmt.Sprintf("An event with the name '%s' already exists in your account", eventRequest.Name),
				Error:   err.Error(),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(common.Response{
			Status:  "error",
			Message: "Failed to create event",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(common.Response{
		Status:  "success",
		Message: "Event created successfully",
		Data:    result,
	})
}

// GetEventDDD demonstrates DDD-compliant event retrieval
func (h *EventHandlerDDD) GetEventDDD(c *fiber.Ctx) error {
	eventId, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.Response{
			Status:  "error",
			Message: "Invalid event ID",
			Error:   err.Error(),
		})
	}

	userId := auth.GetUserId(c)

	query := application.GetEventQuery{
		EventID: eventId,
		UserID:  userId,
	}

	event, err := h.getService.Execute(c.Context(), query)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.Response{
			Status:  "error",
			Message: "Failed to fetch event",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(common.Response{
		Status:  "success",
		Message: "Event fetched successfully",
		Data:    event,
	})
}

// GetEventsDDD demonstrates DDD-compliant event listing
func (h *EventHandlerDDD) GetEventsDDD(c *fiber.Ctx) error {
	queryStr := string(c.Request().URI().QueryString())
	parsedModel := pager.ParseQueryString(queryStr)
	queryModel := EventQueryModel(parsedModel)

	if queryModel.Filter["status"] == "all" {
		delete(queryModel.Filter, "status")
	}

	queryModel.Size = 10
	userId := auth.GetUserId(c)

	query := application.GetEventsQuery{
		UserID: userId,
		Search: queryModel.Search,
		Filter: queryModel.Filter,
		Range:  queryModel.Range,
		Page:   queryModel.Page,
		Size:   queryModel.Size,
	}

	events, err := h.getEventsService.Execute(c.Context(), query)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.Response{
			Status:  "error",
			Message: "Failed to fetch events",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(common.Response{
		Status:  "success",
		Message: "Events fetched successfully",
		Data:    events,
	})
}

// UpdateEventDDD demonstrates DDD-compliant event update
func (h *EventHandlerDDD) UpdateEventDDD(c *fiber.Ctx) error {
	eventId, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.Response{
			Status:  "error",
			Message: "Invalid event ID",
		})
	}

	var eventRequest CreateEventRequest
	if err := c.BodyParser(&eventRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.Response{
			Status:  "error",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	user := auth.GetUser(c)

	cmd := application.UpdateEventCommand{
		EventID:         eventId,
		UserID:          user.Id,
		Name:            eventRequest.Name,
		Description:     eventRequest.Description,
		StartDate:       eventRequest.StartDate,
		EndDate:         eventRequest.EndDate,
		Location:        eventRequest.Location,
		ModeType:        string(eventRequest.ModeType),
		DurationType:    string(eventRequest.DurationType),
		CategoryId:      eventRequest.CategoryId,
		OtherCategory:   eventRequest.OtherCategory,
	}

	event, err := h.updateService.Execute(c.Context(), cmd)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.Response{
			Status:  "error",
			Message: "Failed to update event",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(common.Response{
		Status:  "success",
		Message: "Event updated successfully",
		Data:    event,
	})
}
