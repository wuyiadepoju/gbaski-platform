package registration

import (
	"strings"

	"github.com/gbaski/gbaski-ext/auth"
	"github.com/gbaski/gbaski-platform/internal/common"
	"github.com/gbaski/gbaski-platform/internal/registration/adapters"
	"github.com/gbaski/gbaski-platform/internal/registration/application"
	"github.com/gbaski/gbaski-platform/internal/registration/contracts"
	"github.com/gbaski/gbaski-platform/internal/registration/domain"
	"github.com/gbaski/gbaski-platform/internal/registration/repo"
	"github.com/gbaski/gbaski-shared/pager"
	redisDB "github.com/gbaski/gbaski-shared/redis"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Response = common.Response

type RegistrationHandler struct {
	getRegistrationsService     *application.GetRegistrationsService
	getAttendeesService         *application.GetAttendeesService
	getRegistrationStatsService *application.GetRegistrationStatsService
	checkInService              *application.CheckInService
	agentService                *application.AgentService
	lookupRegistrationService   *application.LookupRegistrationService
}

func NewRegistrationHandler() *RegistrationHandler {
	registrationRepo := repo.NewRegistrationRepositoryImpl()
	mailService := adapters.NewMailServiceImpl()

	// Redis client from shared/redis
	var agentStore contracts.CheckInAgentStore
	redisClient := redisDB.New()
	if redisClient != nil {
		agentStore = adapters.NewCheckInAgentStoreImpl(redisClient.Client)
	}

	return &RegistrationHandler{
		getRegistrationsService:     application.NewGetRegistrationsService(registrationRepo),
		getAttendeesService:         application.NewGetAttendeesService(registrationRepo),
		getRegistrationStatsService: application.NewGetRegistrationStatsService(registrationRepo),
		checkInService:              application.NewCheckInService(registrationRepo, mailService),
		agentService:                application.NewAgentService(registrationRepo, agentStore),
		lookupRegistrationService:   application.NewLookupRegistrationService(registrationRepo),
	}
}

func (s *RegistrationHandler) GetTicketRegistrations(c *fiber.Ctx) error {
	formID, _ := uuid.Parse(c.Params("formId"))
	userID := auth.GetUserId(c)

	queryStr := string(c.Request().URI().QueryString())
	parsedModel := pager.ParseQueryString(queryStr)
	queryModel := domain.RegistrationQueryModel{
		Search: parsedModel.Search,
		Filter: parsedModel.Filter,
		Range:  parsedModel.Range,
		Page:   parsedModel.Page,
		Size:   10,
	}

	if queryModel.Filter["reg_status"] == "all" {
		delete(queryModel.Filter, "reg_status")
	}

	registrations, err := s.getRegistrationsService.Execute(c.Context(), application.GetRegistrationsQuery{
		Query:  queryModel,
		FormID: formID,
		UserID: userID,
	})

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to fetch registrations",
			Error:   err.Error(),
		})
	}

	dtos := make([]application.RegistrationDTO, len(registrations.Items))
	for i, item := range registrations.Items {
		dtos[i] = application.ToRegistrationDTO(item)
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Registrations fetched successfully",
		Data: map[string]any{
			"items": dtos,
			"page":  registrations.Page,
			"size":  registrations.Size,
			"total": registrations.Total,
		},
	})
}

func (s *RegistrationHandler) GetTicketAttendees(c *fiber.Ctx) error {
	formID, _ := uuid.Parse(c.Params("formId"))
	userID := auth.GetUserId(c)

	queryStr := string(c.Request().URI().QueryString())
	parsedModel := pager.ParseQueryString(queryStr)
	queryModel := domain.RegistrationQueryModel{
		Search: parsedModel.Search,
		Filter: parsedModel.Filter,
		Range:  parsedModel.Range,
		Page:   parsedModel.Page,
		Size:   10,
	}

	attendees, err := s.getAttendeesService.Execute(c.Context(), application.GetAttendeesQuery{
		Query:  queryModel,
		FormID: formID,
		UserID: userID,
	})

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to fetch registration ticket attendees",
			Error:   err.Error(),
		})
	}

	dtos := make([]application.AttendeeDTO, len(attendees.Items))
	for i, item := range attendees.Items {
		dtos[i] = application.ToAttendeeDTO(item)
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Registration ticket attendees fetched successfully",
		Data: map[string]any{
			"items": dtos,
			"page":  attendees.Page,
			"size":  attendees.Size,
			"total": attendees.Total,
		},
	})
}

func (s *RegistrationHandler) UpdateTicketRegistrationStatus(c *fiber.Ctx) error {
	var request struct {
		RegRef string `json:"regRef"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	err := s.checkInService.Execute(c.Context(), application.CheckInCommand{
		RegRef:    request.RegRef,
		AgentName: "System Agent",
	})

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

func (s *RegistrationHandler) GetTicketRegistrationStats(c *fiber.Ctx) error {
	formID, _ := uuid.Parse(c.Params("formId"))
	userID := auth.GetUserId(c)

	stats, err := s.getRegistrationStatsService.Execute(c.Context(), application.GetRegistrationStatsQuery{
		FormID: formID,
		UserID: userID,
	})

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

func (s *RegistrationHandler) GetTicketCheckInAgents(c *fiber.Ctx) error {
	formID, _ := uuid.Parse(c.Params("formId"))

	agents, err := s.agentService.GetAgents(c.Context(), application.GetAgentsQuery{FormID: formID})

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

func (s *RegistrationHandler) CreateTicketCheckInAgent(c *fiber.Ctx) error {
	formID, _ := uuid.Parse(c.Params("formId"))
	var request struct {
		Name string `json:"name"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	agents, err := s.agentService.CreateAgent(c.Context(), application.CreateAgentCommand{
		FormID: formID,
		Name:   request.Name,
	})

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

func (s *RegistrationHandler) DeleteTicketCheckInAgent(c *fiber.Ctx) error {
	formID, _ := uuid.Parse(c.Params("formId"))
	token := c.Params("token")

	agents, err := s.agentService.DeleteAgent(c.Context(), application.DeleteAgentCommand{
		FormID: formID,
		Token:  token,
	})

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

func (s *RegistrationHandler) GetTicketCheckInAgent(c *fiber.Ctx) error {
	state := c.Params("state")

	agent, err := s.agentService.GetAgentByToken(c.Context(), application.GetAgentByTokenQuery{Token: state})

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

func (s *RegistrationHandler) CheckInRegistration(c *fiber.Ctx) error {
	regRef := c.Query("regRef")
	token := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")

	agent, err := s.agentService.GetAgentByToken(c.Context(), application.GetAgentByTokenQuery{Token: token})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to fetch check-in agent",
			Error:   err.Error(),
		})
	}

	err = s.checkInService.Execute(c.Context(), application.CheckInCommand{
		RegRef:    regRef,
		AgentName: agent.Name,
	})

	if err != nil {
		status := fiber.StatusBadRequest
		if err == domain.ErrAttendeeNotFound {
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(Response{
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

func (s *RegistrationHandler) LookupRegistration(c *fiber.Ctx) error {
	regRef := c.Query("regRef")
	token := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")

	agent, err := s.agentService.GetAgentByToken(c.Context(), application.GetAgentByTokenQuery{Token: token})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to fetch check-in agent",
			Error:   err.Error(),
		})
	}

	registration, err := s.lookupRegistrationService.Execute(c.Context(), application.LookupRegistrationQuery{
		FormID: agent.FormID,
		RegRef: regRef,
	})

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
