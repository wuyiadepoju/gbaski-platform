package event

import (
	"fmt"
	"strings"

	"github.com/gbaski/gbaski-ext/auth"
	"github.com/gbaski/gbaski-ext/common"
	"github.com/gbaski/gbaski-ext/job"
	"github.com/gbaski/gbaski-platform/internal/event/adapters"
	"github.com/gbaski/gbaski-platform/internal/event/application"
	"github.com/gbaski/gbaski-platform/internal/event/contracts"
	"github.com/gbaski/gbaski-platform/internal/event/repo"
	formapp "github.com/gbaski/gbaski-platform/internal/form/application"
	formrepo "github.com/gbaski/gbaski-platform/internal/form/repo"
	"github.com/gbaski/gbaski-shared/pager"
	validators "github.com/gbaski/gbaski-shared/validator"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Response type alias
type Response = common.Response

// EventHandler is the DDD-compliant event handler
type EventHandler struct {
	createService          *application.CreateEventService
	getService             *application.GetEventService
	getEventsService       *application.GetEventsService
	getEventsReportService *application.GetEventsReportService
	updateService          *application.UpdateEventService
	updateStatusService    *application.UpdateEventStatusService
	deleteService          *application.DeleteEventService
	updateImageService     *application.UpdateEventImageService
	updateVideoService     *application.UpdateEventVideoService
	categoryRepo           contracts.EventCategoryRepository
	eventRepo              contracts.EventRepository
	formCreateService      *formapp.CreateFormService
	formGetService         *formapp.GetFormsService
}

// NewEventHandler creates a new DDD-compliant event handler
// If db is nil, repositories will use BaseRepository's default DB connection
func NewEventHandler(db *sqlx.DB) *EventHandler {

	eventRepo := repo.NewEventRepositoryImpl(db)
	categoryRepo := repo.NewEventCategoryRepositoryImpl(db)

	eventBus := adapters.NewEventBusImpl()
	webmasterService := adapters.NewWebmasterServiceImpl()

	emailScheduler := adapters.NewEmailSchedulerImpl(
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
	getEventsReportService := application.NewGetEventsReportService(eventRepo)

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

	// Initialize form services using DDD form repository and shared event bus
	formRepository := formrepo.NewFormRepositoryImpl(db)
	formCreateService := formapp.NewCreateFormService(formRepository, eventBus)
	formGetService := formapp.NewGetFormsService(formRepository)

	return &EventHandler{
		createService:          createService,
		getService:             getService,
		getEventsService:       getEventsService,
		getEventsReportService: getEventsReportService,
		updateService:          updateService,
		updateStatusService:    updateStatusService,
		deleteService:          deleteService,
		updateImageService:     updateImageService,
		updateVideoService:     updateVideoService,
		categoryRepo:           categoryRepo,
		eventRepo:              eventRepo,
		formCreateService:      formCreateService,
		formGetService:         formGetService,
	}
}

// @Summary Get all events
// @Description Retrieve a paginated list of events for the authenticated user with optional filtering and search
// @Tags events
// @Accept json
// @Produce json
// @Param search query string false "Search query"
// @Param filter query string false "Filter parameters"
// @Param range query string false "Range parameters"
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(10)
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /events [get]
func (h *EventHandler) GetEvents(c *fiber.Ctx) error {
	queryStr := string(c.Request().URI().QueryString())
	parsedModel := pager.ParseQueryString(queryStr)
	queryModel := application.EventQueryModel(parsedModel)

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
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to fetch events",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Events fetched successfully",
		Data:    events,
	})
}

// @Summary Get event by ID
// @Description Retrieve a single event by its ID
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID (UUID)"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /events/{id} [get]
func (h *EventHandler) GetEvent(c *fiber.Ctx) error {
	eventId, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
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
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to fetch event",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Event fetched successfully",
		Data:    event,
	})
}

// @Summary Create new event
// @Description Create a new event for the authenticated user
// @Tags events
// @Accept json
// @Produce json
// @Param request body application.CreateEventRequest true "Event creation details"
// @Success 201 {object} Response
// @Failure 400 {object} Response
// @Failure 409 {object} Response
// @Security Bearer
// @Router /events [post]
func (h *EventHandler) CreateEvent(c *fiber.Ctx) error {
	var eventRequest application.CreateEventRequest

	if err := c.BodyParser(&eventRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	// Validate request
	isValidData := validators.Validate(eventRequest)
	if !isValidData {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: validators.ErrorMessage(),
		})
	}

	user := auth.GetUser(c)

	// Map request to command
	cmd := application.CreateEventCommand{
		UserID:           user.Id,
		Name:             eventRequest.Name,
		Description:      eventRequest.Description,
		StartDate:        eventRequest.StartDate,
		EndDate:          eventRequest.EndDate,
		Location:         eventRequest.Location,
		ModeType:         eventRequest.ModeType,
		Payment:          eventRequest.Payment,
		AccessType:       eventRequest.AccessType,
		DurationType:     eventRequest.DurationType,
		RegistrationType: eventRequest.RegType,
		CategoryId:       eventRequest.CategoryId,
		OtherCategory:    eventRequest.OtherCategory,
		BrandName:        user.Name,
	}

	// Execute use case
	result, err := h.createService.Execute(c.Context(), cmd)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "duplicate") && strings.Contains(errStr, "events_name_user_id_unique") {
			return c.Status(fiber.StatusConflict).JSON(Response{
				Status:  "error",
				Message: fmt.Sprintf("An event with the name '%s' already exists in your account", eventRequest.Name),
				Error:   err.Error(),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to create event",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(Response{
		Status:  "success",
		Message: "Event created successfully",
		Data:    result,
	})
}

// @Summary Update event
// @Description Update an existing event by ID
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID (UUID)"
// @Param request body application.CreateEventRequest true "Event update details"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /events/{id} [put]
func (h *EventHandler) UpdateEvent(c *fiber.Ctx) error {
	eventId, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Invalid event ID",
		})
	}

	var eventRequest application.CreateEventRequest
	if err := c.BodyParser(&eventRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	user := auth.GetUser(c)

	cmd := application.UpdateEventCommand{
		EventID:       eventId,
		UserID:        user.Id,
		Name:          eventRequest.Name,
		Description:   eventRequest.Description,
		StartDate:     eventRequest.StartDate,
		EndDate:       eventRequest.EndDate,
		Location:      eventRequest.Location,
		ModeType:      eventRequest.ModeType,
		DurationType:  eventRequest.DurationType,
		CategoryId:    eventRequest.CategoryId,
		OtherCategory: eventRequest.OtherCategory,
	}

	event, err := h.updateService.Execute(c.Context(), cmd)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to update event",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Event updated successfully",
		Data:    event,
	})
}

// @Summary Update event status
// @Description Update the status of an existing event
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID (UUID)"
// @Param request body application.StatusRequest true "Status update request"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /events/{id}/status [patch]
func (h *EventHandler) UpdateEventStatus(c *fiber.Ctx) error {
	eventId, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Invalid event ID",
		})
	}

	var statusRequest application.StatusRequest
	if err := c.BodyParser(&statusRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	userId := auth.GetUserId(c)

	cmd := application.UpdateEventStatusCommand{
		EventID: eventId,
		UserID:  userId,
		Status:  statusRequest.Status.Value(),
	}

	event, err := h.updateStatusService.Execute(c.Context(), cmd)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to update event status",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Event status updated successfully",
		Data:    event,
	})
}

// @Summary Delete event
// @Description Delete an event by ID
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID (UUID)"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /events/{id} [delete]
func (h *EventHandler) DeleteEvent(c *fiber.Ctx) error {
	eventId, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Invalid event ID",
		})
	}

	userId := auth.GetUserId(c)

	cmd := application.DeleteEventCommand{
		EventID: eventId,
		UserID:  userId,
	}

	err = h.deleteService.Execute(c.Context(), cmd)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to delete event",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Event deleted successfully",
		Data:    eventId,
	})
}

// @Summary Update event image
// @Description Update the image URL for an event
// @Tags events
// @Accept json
// @Produce json
// @Param request body application.UpdateEventImageRequest true "Event image update request"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /events/image [post]
func (h *EventHandler) UpdateEventImage(c *fiber.Ctx) error {
	var updateEventImageRequest application.UpdateEventImageRequest

	if err := c.BodyParser(&updateEventImageRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	if updateEventImageRequest.ImageURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Image URL is required",
		})
	}

	cmd := application.UpdateEventImageCommand{
		EventID:  updateEventImageRequest.EventId,
		ImageURL: updateEventImageRequest.ImageURL,
	}

	err := h.updateImageService.Execute(c.Context(), cmd)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to update event image",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Event image updated successfully",
		Data:    fiber.Map{"eventId": updateEventImageRequest.EventId},
	})
}

// @Summary Update event video
// @Description Update the video URL for an event
// @Tags events
// @Accept json
// @Produce json
// @Param request body application.UpdateEventVideoRequest true "Event video update request"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /events/video [post]
func (h *EventHandler) UpdateEventVideo(c *fiber.Ctx) error {
	var updateEventVideoRequest application.UpdateEventVideoRequest

	if err := c.BodyParser(&updateEventVideoRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	cmd := application.UpdateEventVideoCommand{
		EventID:  updateEventVideoRequest.EventId,
		VideoURL: updateEventVideoRequest.VideoURL,
	}

	err := h.updateVideoService.Execute(c.Context(), cmd)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to update event video",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Event video updated successfully",
		Data:    fiber.Map{"eventId": updateEventVideoRequest.EventId},
	})
}

// @Summary Get event categories
// @Description Retrieve a list of all available event categories
// @Tags events
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /events/categories [get]
func (h *EventHandler) GetEventCategories(c *fiber.Ctx) error {
	categories, err := h.categoryRepo.FindAll()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to fetch event categories",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Event categories fetched successfully",
		Data:    categories,
	})
}

// @Summary Get event statistics
// @Description Retrieve statistics for a specific event
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID (UUID)"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /events/{id}/stats [get]
func (h *EventHandler) GetEventStats(c *fiber.Ctx) error {
	// TODO: Create GetEventStatsService in application layer
	// For now, return not implemented
	return c.Status(fiber.StatusNotImplemented).JSON(Response{
		Status:  "error",
		Message: "Event stats feature needs to be migrated to DDD application service",
	})
}

// @Summary Get event forms
// @Description Retrieve all registration forms for the authenticated user's events
// @Tags events
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /registrations/forms [get]
func (h *EventHandler) GetEventForms(c *fiber.Ctx) error {
	// For now, expect an eventId query param to fetch forms for a specific event
	eventIdStr := c.Query("eventId")
	if eventIdStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "eventId query parameter is required",
		})
	}

	eventId, err := uuid.Parse(eventIdStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Invalid event ID",
			Error:   err.Error(),
		})
	}

	query := formapp.GetFormsQuery{
		EventID: eventId,
	}

	forms, err := h.formGetService.Execute(c.Context(), query)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to fetch event forms",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Event forms fetched successfully",
		Data:    forms,
	})
}

// @Summary Create registration form
// @Description Create a new registration form for an event
// @Tags events
// @Accept json
// @Produce json
// @Param request body FormRequest true "Form creation details"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /registrations/forms [post]
func (h *EventHandler) CreateForm(c *fiber.Ctx) error {
	var formRequest formapp.CreateFormRequest

	if err := c.BodyParser(&formRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	userId := auth.GetUserId(c)

	cmd := formapp.CreateFormCommand{
		UserID:            userId,
		CreateFormRequest: formRequest,
	}

	formId, err := h.formCreateService.Execute(c.Context(), cmd)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to create form",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Form created successfully",
		Data:    formId,
	})
}

// @Summary Get events report
// @Description Retrieve a report of all events for the authenticated user
// @Tags events
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /events/report [get]
func (h *EventHandler) GetEventsReport(c *fiber.Ctx) error {
	userId := auth.GetUserId(c)

	query := application.GetEventsReportQuery{
		UserID: userId,
	}

	report, err := h.getEventsReportService.Execute(c.Context(), query)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to fetch events report",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Events report fetched successfully",
		Data:    report,
	})
}
