package main

import (
	"context"

	"github.com/gbaski/gbaski-event/pkg/payment"
	"github.com/gbaski/gbaski-event/pkg/thirdparty/google"
	"github.com/gbaski/gbaski-ext/auth"
	"github.com/gbaski/gbaski-ext/common"
	"github.com/gbaski/gbaski-ext/upload"
	"github.com/gbaski/gbaski-platform/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
)

var paymentHandler = payment.NewPaymentHandler()

// Type references for Swagger documentation (used in annotations)
var (
	_ auth.LoginRequest
	_ auth.RegisterRequest
	_ auth.RefreshTokenRequest
	_ auth.LoginResponse
	_ auth.Response
	_ common.Response
)

// Auth handler wrappers with Swagger annotations

// @Summary Login user
// @Description Authenticate a user and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body auth.LoginRequest true "Login credentials"
// @Success 200 {object} auth.LoginResponse
// @Failure 400 {object} auth.Response
// @Failure 401 {object} auth.Response
// @Failure 500 {object} auth.Response
// @Router /auth/login [post]
func handleLogin(c *fiber.Ctx) error {
	return authHandler.Login(c)
}

// @Summary Login user with OTP
// @Description Authenticate a user using OTP code sent to their email
// @Tags auth
// @Accept json
// @Produce json
// @Param otp_email query string true "Email address"
// @Param otp_code query string false "OTP code (required for verification)"
// @Success 200 {object} auth.Response
// @Failure 400 {object} auth.Response
// @Failure 404 {object} auth.Response
// @Failure 500 {object} auth.Response
// @Router /auth/otp-login [post]
func handleOtpLogin(c *fiber.Ctx) error {
	return authHandler.OtpLogin(c)
}

// @Summary Register new user
// @Description Register a new user in the system
// @Tags auth
// @Accept json
// @Produce json
// @Param request body auth.RegisterRequest true "User registration details"
// @Success 200 {object} auth.LoginResponse
// @Failure 400 {object} auth.Response
// @Failure 409 {object} auth.Response
// @Failure 500 {object} auth.Response
// @Router /auth/register [post]
func handleRegister(c *fiber.Ctx) error {
	return authHandler.Register(c)
}

// @Summary Check if a name is available
// @Description Check if a name is available for registration
// @Tags auth
// @Accept json
// @Produce json
// @Param name path string true "Name to check"
// @Success 200 {object} auth.Response
// @Failure 400 {object} auth.Response
// @Failure 404 {object} auth.Response
// @Failure 500 {object} auth.Response
// @Router /auth/available-name/{name} [get]
func handleCheckAvailableName(c *fiber.Ctx) error {
	return authHandler.CheckAvailableName(c)
}

// @Summary Refresh access token
// @Description Get a new access token using a valid refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body auth.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} auth.Response
// @Failure 400 {object} auth.Response
// @Failure 401 {object} auth.Response
// @Failure 500 {object} auth.Response
// @Router /auth/refresh-token [post]
func handleRefreshToken(c *fiber.Ctx) error {
	return authHandler.RefreshToken(c)
}

// @Summary Change user phone number
// @Description Update the phone number for the authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]any true "Phone number change request" example({"phone":"+1234567890"})
// @Success 200 {object} auth.Response
// @Failure 400 {object} auth.Response
// @Failure 401 {object} auth.Response
// @Failure 500 {object} auth.Response
// @Security Bearer
// @Router /settings/change-phone [post]
func handleChangePhone(c *fiber.Ctx) error {
	return authHandler.ChangePhone(c)
}

// @Summary Change user password
// @Description Update the password for the authenticated user using OTP verification
// @Tags auth
// @Accept json
// @Produce json
// @Param otp_code query string false "OTP code for verification"
// @Param request body map[string]any true "Password change request" example({"password":"newpassword123"})
// @Success 200 {object} auth.Response
// @Failure 400 {object} auth.Response
// @Failure 401 {object} auth.Response
// @Failure 500 {object} auth.Response
// @Security Bearer
// @Router /settings/change-password [post]
func handleChangePassword(c *fiber.Ctx) error {
	return authHandler.ChangePassword(c)
}

// @Summary Get signed URL for file upload
// @Description Generate a pre-signed URL for uploading files to S3
// @Tags upload
// @Accept json
// @Produce json
// @Param objectName query string true "Object name for the file"
// @Param contentType query string true "Content type of the file (e.g., image/jpeg, video/mp4)"
// @Param bucketName query string false "S3 bucket name" default(gbaski-storage)
// @Success 200 {object} common.Response
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security Bearer
// @Router /upload/signed-url [get]
func handleGetSignedURL(c *fiber.Ctx) error {
	return upload.GetSignedURL(c)
}

func SetupRoutes(app *fiber.App) *fiber.App {

	// Custom error handler that logs to Loki
	app.Use(middleware.ErrorHandler)

	app.Use(logger.New())

	// Custom recover middleware that logs to Loki
	app.Use(recover.New(middleware.RecoverConfig()))

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "https://host.gbaski.web:6200,http://localhost:6200,http://host.gbaski.web:6200,https://gbaski.app,https://host.gbaski.app,https://platform.gbaski.app,http://*.gbaski.web,http://host.gbaski.web,http://api.gbaski.web",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-API-Key, X-Brand-Name",
		ExposeHeaders:    "Content-Type, Authorization",
		AllowCredentials: true,
		MaxAge:           86400,
	}))

	// Handle OPTIONS preflight requests
	app.Options("/*", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// Robots.txt - Block all crawling
	app.Get("/robots.txt", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/plain")
		return c.SendString("User-agent: *\nDisallow: /\n")
	})

	// Security monitoring: Log suspicious query parameters and potential security threats
	app.All("/", middleware.SecurityMonitor, func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusOK)
	})

	app.Get("/google/webmaster/list-properties", func(c *fiber.Ctx) error {

		webmasterService, err := google.NewWebmasterService(context.Background())

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to initialize webmaster service",
				"status":  "error",
			})
		}

		properties, err := webmasterService.ListSites()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to list properties",
				"status":  "error",
			})
		}

		return c.JSON(fiber.Map{
			"message":    "Properties listed successfully",
			"status":     "success",
			"properties": properties,
		})
	})

	app.Get("/google/webmaster/inspect-url/:url", func(c *fiber.Ctx) error {

		webmasterService, err := google.NewWebmasterService(context.Background())

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to initialize webmaster service",
				"status":  "error",
			})
		}

		url := c.Params("url")

		result, err := webmasterService.InspectURL(url)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to inspect URL",
				"status":  "error",
			})
		}

		return c.JSON(fiber.Map{
			"message": "URL inspected successfully",
			"status":  "success",
			"result":  result,
		})

	})

	app.Get("/auth/available-name/:name", handleCheckAvailableName)
	app.Post("/auth/login", handleLogin)
	app.Post("/auth/otp-login", handleOtpLogin)
	app.Post("/auth/register", handleRegister)
	app.Post("/auth/refresh-token", handleRefreshToken)

	app.Get("/registrations/lookup", registrationHandler.LookupRegistration)
	app.Get("/registrations/check-in", registrationHandler.CheckInRegistration)

	app.Use(middleware.Protected)
	// app.Use(middleware.AccessControl)

	// Apply request context logging middleware to all protected routes
	// This automatically logs user context and route info to Loki
	app.Use(middleware.LogRequestContext)

	app.Get("/protected", func(c *fiber.Ctx) error {

		return c.SendString("Protected")
	})

	app.Get("/payment/ref", paymentHandler.GetPaymentRef)
	// app.Get("/payment/order-summary/:regRef", paymentHandler.DownloadOrderSummary)
	app.Use("/payment/*", paymentHandler.Proxy)

	// Event routes
	app.Get("/events", eventHandler.GetEvents)
	app.Get("/events/categories", eventHandler.GetEventCategories)
	app.Post("/events", eventHandler.CreateEvent)
	app.Get("/events/:id/stats", eventHandler.GetEventStats)
	app.Post("/events/image", eventHandler.UpdateEventImage)
	app.Post("/events/video", eventHandler.UpdateEventVideo)
	app.Get("/events/report", eventHandler.GetEventsReport)
	app.Get("/events/:id", eventHandler.GetEvent)
	app.Put("/events/:id", eventHandler.UpdateEvent)
	app.Patch("/events/:id/status", eventHandler.UpdateEventStatus)
	app.Delete("/events/:id", eventHandler.DeleteEvent)

	app.Post("/registrations/forms", eventHandler.CreateForm)
	app.Get("/registrations/forms", eventHandler.GetEventForms)

	ticketRegistration := app.Group("/registrations/ticket/:formId")
	ticketRegistration.Get("/", registrationHandler.GetTicketRegistrations)
	ticketRegistration.Get("/check-in-agents", registrationHandler.GetTicketCheckInAgents)
	ticketRegistration.Post("/check-in-agents", registrationHandler.CreateTicketCheckInAgent)
	ticketRegistration.Delete("/check-in-agents/:token", registrationHandler.DeleteTicketCheckInAgent)
	ticketRegistration.Get("/attendees", registrationHandler.GetTicketAttendees)
	ticketRegistration.Get("/stats", registrationHandler.GetTicketRegistrationStats)
	ticketRegistration.Put("/status", registrationHandler.UpdateTicketRegistrationStatus)

	app.Get("/payout/account", payoutHandler.GetPayoutAccount)
	app.Get("/payout/banks", payoutHandler.GetPayoutBanks)
	app.Post("/payout/create-account", payoutHandler.CreatePayoutAccount)
	app.Get("/payout/history/:formId", payoutHandler.GetPayoutHistory)
	app.Get("/payout/summary/:formId", payoutHandler.GetPayoutSummary)
	app.Post("/payout/change-type", payoutHandler.ChangePayoutType)
	app.Post("/payout/payment-provider", payoutHandler.CreatePaymentPayout)
	app.Get("/payout/payment-provider", payoutHandler.GetPaymentPayout)
	app.Get("/payout/total-payout/:formId", payoutHandler.GetTotalPayout)

	app.Get("/wallet/balance", walletHandler.GetBalance)

	app.Post("/settings/change-phone", handleChangePhone)
	app.Post("/settings/change-password", handleChangePassword)

	app.Get("/upload/signed-url", handleGetSignedURL)

	// Account routes
	// accountGroup := app.Group("/accounts")

	// accountGroup.Get("/profile", account.GetProfile)
	// accountGroup.Put("/profile", account.UpdateProfile)
	// accountGroup.Put("/change-password", account.ChangePassword)

	// Swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault)
	app.Get("/swagger/*", swagger.New(swagger.Config{
		URL:          "/docs/swagger.json",
		DeepLinking:  false,
		DocExpansion: "none",
	}))

	return app
}
