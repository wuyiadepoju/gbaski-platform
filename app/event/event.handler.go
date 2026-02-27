package event

import (
	"fmt"
	"strings"

	"github.com/gbaski/gbaski-platform/app/form"
	"github.com/gofiber/fiber/v2"

	"github.com/google/uuid"

	"github.com/gbaski/gbaski-ext/auth"
	"github.com/gbaski/gbaski-ext/job"
	"github.com/gbaski/gbaski-ext/mailcoach"
	"github.com/gbaski/gbaski-shared/pager"
	validators "github.com/gbaski/gbaski-shared/validator"
)

// Type references for Swagger documentation (used in annotations)
var (
	_ Response
	_ CreateEventRequest
	_ EventModel
	_ StatusRequest
	_ UpdateEventImageRequest
	_ UpdateEventVideoRequest
	_ FormRequest
)

// Handler service
type EventHandler struct {
	*Service
}

func NewEventHandler() *EventHandler {

	newService := NewService()
	newService.formHandler = form.NewFormHandler()
	newService.jobService = job.NewService()
	newService.mailcoach = mailcoach.New()
	newService.authHandler = auth.NewAuthHandler()
	return &EventHandler{
		newService,
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
// GetEvents retrieves all events
func (s *EventHandler) GetEvents(c *fiber.Ctx) error {

	queryStr := string(c.Request().URI().QueryString())

	parsedModel := pager.ParseQueryString(queryStr)

	queryModel := EventQueryModel(parsedModel)

	if queryModel.Filter["status"] == "all" {
		delete(queryModel.Filter, "status")
	}

	queryModel.Size = 10

	userId := auth.GetUserId(c)

	events, err := s.getEvents(queryModel, userId)

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
// GetEvent retrieves a single event by ID
func (s *EventHandler) GetEvent(c *fiber.Ctx) error {

	eventId, _ := uuid.Parse(c.Params("id"))

	eventDetail, err := s.getEvent(eventId)

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
		Data:    &eventDetail,
	})
}

// @Summary Create new event
// @Description Create a new event for the authenticated user
// @Tags events
// @Accept json
// @Produce json
// @Param request body CreateEventRequest true "Event creation details"
// @Success 201 {object} Response
// @Failure 400 {object} Response
// @Failure 409 {object} Response
// @Security Bearer
// @Router /events [post]
// CreateEvent creates a new event
func (s *EventHandler) CreateEvent(c *fiber.Ctx) error {

	var eventRequest CreateEventRequest

	if err := c.BodyParser(&eventRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"status":  "error",
			"error":   err.Error(),
		})
	}

	isValidData := validators.Validate(eventRequest)

	if !isValidData {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: validators.ErrorMessage(),
		})
	}

	user := auth.GetUser(c)

	eventModel := EventModel{
		CreateEventRequest: eventRequest,
		UserId:             user.Id,
		BrandName:          user.Name,
	}

	createEventResponse, err := s.createUpdateEvent(eventModel, user)

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
		Data:    createEventResponse,
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
func (s *EventHandler) GetEventStats(c *fiber.Ctx) error {

	eventId, _ := uuid.Parse(c.Params("id"))
	userId := auth.GetUserId(c)

	stats, err := s.getEventStats(eventId, userId)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to fetch event stats",
			Error:   err.Error(),
		})
	}

	return c.JSON(Response{
		Status:  "success",
		Message: "Event stats fetched successfully",
		Data:    stats,
	})
}

// @Summary Update event
// @Description Update an existing event by ID
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID (UUID)"
// @Param request body CreateEventRequest true "Event update details"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /events/{id} [put]
// UpdateEvent updates an existing event
func (s *EventHandler) UpdateEvent(c *fiber.Ctx) error {

	eventId, _ := uuid.Parse(c.Params("id"))
	var eventRequest CreateEventRequest

	if err := c.BodyParser(&eventRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	user := auth.GetUser(c)

	eventModel := EventModel{
		Id:                 &eventId,
		CreateEventRequest: eventRequest,
		UserId:             user.Id,
	}

	_, err := s.createUpdateEvent(eventModel, user)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to update event",
			Error:   err.Error(),
		})
	}

	event, err := s.getEvent(eventId)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to get event",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Event updated successfully",
		Data:    event,
	})
}

type StatusRequest struct {
	Status EventStatus
}

// @Summary Update event status
// @Description Update the status of an existing event
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID (UUID)"
// @Param request body StatusRequest true "Status update request"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /events/{id}/status [patch]
func (s *EventHandler) UpdateEventStatus(c *fiber.Ctx) error {

	eventId, _ := uuid.Parse(c.Params("id"))

	var statusRequest StatusRequest

	if err := c.BodyParser(&statusRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	event, err := s.updateEventStatus(eventId, statusRequest.Status)

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
// DeleteEvent deletes an event
func (s *EventHandler) DeleteEvent(c *fiber.Ctx) error {

	eventId, _ := uuid.Parse(c.Params("id"))

	err := s.deleteEvent(eventId)

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

type UpdateEventImageRequest struct {
	EventId  uuid.UUID `json:"eventId"`
	ImageUrl string    `json:"imageUrl"`
}

// @Summary Update event image
// @Description Update the image URL for an event
// @Tags events
// @Accept json
// @Produce json
// @Param request body UpdateEventImageRequest true "Event image update request"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /events/image [post]
func (s *EventHandler) UpdateEventImage(c *fiber.Ctx) error {

	var updateEventImageRequest UpdateEventImageRequest

	if err := c.BodyParser(&updateEventImageRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	if updateEventImageRequest.ImageUrl == "" {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Image URL is required",
		})
	}

	err := s.updateEventImage(updateEventImageRequest.EventId, updateEventImageRequest.ImageUrl)

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

type UpdateEventVideoRequest struct {
	EventId  uuid.UUID `json:"eventId"`
	VideoUrl string    `json:"videoUrl"`
}

// @Summary Update event video
// @Description Update the video URL for an event
// @Tags events
// @Accept json
// @Produce json
// @Param request body UpdateEventVideoRequest true "Event video update request"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /events/video [post]
func (s *EventHandler) UpdateEventVideo(c *fiber.Ctx) error {

	var updateEventVideoRequest UpdateEventVideoRequest

	if err := c.BodyParser(&updateEventVideoRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	err := s.updateEventVideo(updateEventVideoRequest.EventId, updateEventVideoRequest.VideoUrl)

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
func (s *EventHandler) GetEventCategories(c *fiber.Ctx) error {

	categories, err := s.getEventCategories()

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

// @Summary Get event forms
// @Description Retrieve all registration forms for the authenticated user's events
// @Tags events
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /registrations/forms [get]
func (s *EventHandler) GetEventForms(c *fiber.Ctx) error {

	userId := auth.GetUserId(c)

	forms, err := s.getEventForms(userId)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to fetch forms",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Forms fetched successfully",
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
func (s *EventHandler) CreateForm(c *fiber.Ctx) error {

	var createFormRequest FormRequest

	if err := c.BodyParser(&createFormRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	user := auth.GetUser(c)

	formId, err := s.createForm(createFormRequest, user)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to create form",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Form Published Successfully",
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
func (s *EventHandler) GetEventsReport(c *fiber.Ctx) error {

	userId := auth.GetUserId(c)

	report, err := s.getEventsReport(userId)

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
