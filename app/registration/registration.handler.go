package registration

import (
	"strings"

	"github.com/gbaski/gbaski-ext/auth"
	"github.com/gbaski/gbaski-shared/pager"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Type references for Swagger documentation (used in annotations)
var (
	_ Response
	_ UpdateRegistrationStatusRequest
	_ CreateCheckInAgentRequest
)

type RegistrationHandler struct {
	repo          *Repository
	ticketService *TicketService
	voteService   *VoteService
}

func NewRegistrationHandler() *RegistrationHandler {

	return &RegistrationHandler{
		repo:          NewRepository(),
		ticketService: NewTicketService(),
		voteService:   NewVoteService(),
	}
}

// @Summary Get ticket registrations
// @Description Retrieve a paginated list of ticket registrations for a specific form
// @Tags registrations
// @Accept json
// @Produce json
// @Param formId path string true "Form ID (UUID)"
// @Param search query string false "Search query"
// @Param filter query string false "Filter parameters"
// @Param range query string false "Range parameters"
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(10)
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /registrations/ticket/{formId} [get]
func (s *RegistrationHandler) GetTicketRegistrations(c *fiber.Ctx) error {

	formId, _ := uuid.Parse(c.Params("formId"))

	queryStr := string(c.Request().URI().QueryString())

	parsedModel := pager.ParseQueryString(queryStr)

	queryModel := RegistrationQueryModel(parsedModel)

	if queryModel.Filter["reg_status"] == "all" {
		delete(queryModel.Filter, "reg_status")
	}

	queryModel.Size = 10

	userId := auth.GetUserId(c)

	// fmt.Println(utils.PrettyPrint(queryModel))

	registrations, err := s.ticketService.GetRegistrations(queryModel, formId, userId)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to fetch registrations",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Registrations fetched successfully",
		Data:    registrations,
	})

}

// @Summary Get ticket attendees
// @Description Retrieve a paginated list of ticket attendees for a specific form
// @Tags registrations
// @Accept json
// @Produce json
// @Param formId path string true "Form ID (UUID)"
// @Param search query string false "Search query"
// @Param filter query string false "Filter parameters"
// @Param range query string false "Range parameters"
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(10)
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /registrations/ticket/{formId}/attendees [get]
func (s *RegistrationHandler) GetTicketAttendees(c *fiber.Ctx) error {

	formId, _ := uuid.Parse(c.Params("formId"))

	userId := auth.GetUserId(c)

	queryStr := string(c.Request().URI().QueryString())

	parsedModel := pager.ParseQueryString(queryStr)

	queryModel := RegistrationQueryModel(parsedModel)

	queryModel.Size = 10

	attendees, err := s.ticketService.GetTicketAttendees(queryModel, formId, userId)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to fetch registration ticket attendees",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Registration ticket attendees fetched successfully",
		Data:    attendees,
	})

}

type UpdateRegistrationStatusRequest struct {
	FormId    uuid.UUID          `json:"formId"`
	RegType   string             `json:"regType"`
	RegRef    string             `json:"regRef"`
	RegId     int64              `json:"regId"`
	RegStatus RegistrationStatus `json:"regStatus"`
}

// @Summary Update ticket registration status
// @Description Update the status of a ticket registration (check-in)
// @Tags registrations
// @Accept json
// @Produce json
// @Param formId path string true "Form ID (UUID)"
// @Param request body UpdateRegistrationStatusRequest true "Registration status update request"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /registrations/ticket/{formId}/status [put]
func (s *RegistrationHandler) UpdateTicketRegistrationStatus(c *fiber.Ctx) error {

	var request UpdateRegistrationStatusRequest

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	err := s.ticketService.CheckInTicketRegistration(request.RegRef, "System Agent")

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to check in registration",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Registration checked in successfully",
		Data:    true,
	})
}

// @Summary Get ticket registration statistics
// @Description Retrieve statistics for ticket registrations of a specific form
// @Tags registrations
// @Accept json
// @Produce json
// @Param formId path string true "Form ID (UUID)"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /registrations/ticket/{formId}/stats [get]
func (s *RegistrationHandler) GetTicketRegistrationStats(c *fiber.Ctx) error {

	formId, _ := uuid.Parse(c.Params("formId"))

	userId := auth.GetUserId(c)

	stats, err := s.ticketService.GetRegistrationStats(formId, userId)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to fetch registration stats",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Registration stats fetched successfully",
		Data:    stats,
	})

}

// @Summary Get ticket check-in agents
// @Description Retrieve all check-in agents for a specific form
// @Tags registrations
// @Accept json
// @Produce json
// @Param formId path string true "Form ID (UUID)"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /registrations/ticket/{formId}/check-in-agents [get]
func (s *RegistrationHandler) GetTicketCheckInAgents(c *fiber.Ctx) error {

	formId, _ := uuid.Parse(c.Params("formId"))

	agents, err := s.ticketService.GetCheckInAgents(formId)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to fetch check-in agents",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Check-in agents fetched successfully",
		Data:    agents,
	})

}

type CreateCheckInAgentRequest struct {
	Name string `json:"name"`
}

// @Summary Create ticket check-in agent
// @Description Create a new check-in agent for a specific form
// @Tags registrations
// @Accept json
// @Produce json
// @Param formId path string true "Form ID (UUID)"
// @Param request body CreateCheckInAgentRequest true "Check-in agent creation details"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /registrations/ticket/{formId}/check-in-agents [post]
func (s *RegistrationHandler) CreateTicketCheckInAgent(c *fiber.Ctx) error {

	formId, _ := uuid.Parse(c.Params("formId"))

	var request CreateCheckInAgentRequest

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	agents, err := s.ticketService.CreateCheckInAgent(formId, request.Name)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to create check-in agent",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Check-in agent created successfully",
		Data:    agents,
	})

}

// @Summary Delete ticket check-in agent
// @Description Delete a check-in agent by token for a specific form
// @Tags registrations
// @Accept json
// @Produce json
// @Param formId path string true "Form ID (UUID)"
// @Param token path string true "Agent token"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /registrations/ticket/{formId}/check-in-agents/{token} [delete]
func (s *RegistrationHandler) DeleteTicketCheckInAgent(c *fiber.Ctx) error {

	formId, _ := uuid.Parse(c.Params("formId"))
	token := c.Params("token")

	agents, err := s.ticketService.DeleteCheckInAgent(formId, token)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to delete check-in agent",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Check-in agent deleted successfully",
		Data:    agents,
	})

}

// @Summary Get ticket check-in agent
// @Description Retrieve a check-in agent by state token
// @Tags registrations
// @Accept json
// @Produce json
// @Param state path string true "Agent state token"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /registrations/ticket/check-in-agent/{state} [get]
func (s *RegistrationHandler) GetTicketCheckInAgent(c *fiber.Ctx) error {

	state := c.Params("state")

	agent, err := s.ticketService.GetCheckInAgent(state)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to fetch check-in agent",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Check-in agent fetched successfully",
		Data:    agent,
	})

}

// @Summary Check in registration
// @Description Check in a registration using registration reference and agent token
// @Tags registrations
// @Accept json
// @Produce json
// @Param regRef query string true "Registration reference"
// @Param Authorization header string true "Bearer token (check-in agent token)"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /registrations/check-in [get]
func (s *RegistrationHandler) CheckInRegistration(c *fiber.Ctx) error {

	regRef := c.Query("regRef")

	token := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")

	agent, err := s.ticketService.GetCheckInAgent(token)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to fetch check-in agent",
			Error:   err.Error(),
		})
	}

	err = s.ticketService.CheckInTicketRegistration(regRef, agent.Name)

	if err != nil {

		if err.Error() == "already_checked_in" {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Status:  "error",
				Message: "Already checked in",
				Error:   err.Error(),
			})
		}

		if err.Error() == "regref_not_found" {
			return c.Status(fiber.StatusNotFound).JSON(Response{
				Status:  "error",
				Message: "Registration not found",
				Error:   err.Error(),
			})
		}

		if err.Error() == "agent_not_found" {
			return c.Status(fiber.StatusNotFound).JSON(Response{
				Status:  "error",
				Message: "Agent not found",
				Error:   err.Error(),
			})
		}

		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to check in",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Check in successful",
		Data:    true,
	})

}

// @Summary Lookup registration
// @Description Lookup a registration by registration reference using agent token
// @Tags registrations
// @Accept json
// @Produce json
// @Param regRef query string true "Registration reference"
// @Param Authorization header string true "Bearer token (check-in agent token)"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /registrations/lookup [get]
func (s *RegistrationHandler) LookupRegistration(c *fiber.Ctx) error {

	regRef := c.Query("regRef")

	token := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")

	agent, err := s.ticketService.GetCheckInAgent(token)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to fetch check-in agent",
			Error:   err.Error(),
		})
	}

	registration, err := s.ticketService.LookupRegistration(agent.FormId, regRef)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to lookup registration",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Registration lookup successful",
		Data:    registration,
	})

}
