package payout

import (
	"fmt"

	"github.com/gbaski/gbaski-ext/auth"
	"github.com/gbaski/gbaski-ext/otp"
	"github.com/gbaski/gbaski-platform/app/common"
	"github.com/gbaski/gbaski-shared/pager"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Type references for Swagger documentation (used in annotations)
var (
	_ Response
	_ CreatePayoutAccountRequest
	_ ChangePayoutTypeRequest
	_ CreatePaymentPayoutRequest
)

type PayoutHandler struct {
	*Service
}

func NewPayoutHandler() *PayoutHandler {
	return &PayoutHandler{NewService()}
}

// @Summary Get payout banks
// @Description Retrieve a list of available banks for payout account setup
// @Tags payout
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /payout/banks [get]
func (s *PayoutHandler) GetPayoutBanks(c *fiber.Ctx) error {

	banks, err := s.getBanks()

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to get payout banks",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Payout banks fetched successfully",
		Data:    banks,
	})

}

// @Summary Create payout account
// @Description Create a new payout account for the authenticated user with OTP verification
// @Tags payout
// @Accept json
// @Produce json
// @Param otp_code query string false "OTP code for verification"
// @Param request body CreatePayoutAccountRequest true "Payout account creation details"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /payout/create-account [post]
func (s *PayoutHandler) CreatePayoutAccount(c *fiber.Ctx) error {

	otpCode := c.Query("otp_code", "")

	var data CreatePayoutAccountRequest

	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	email := auth.GetUserEmail(c)

	otpRequest := otp.OtpRequest{
		Key:   fmt.Sprintf("payout_account:%s", email),
		Email: email,
		Code:  otpCode,
	}

	data.UserId = auth.GetUserId(c)
	data.Currency = "NGN"

	payoutAccount, err := s.createPayoutAccount(data, otpRequest)

	if err != nil {

		response := otp.NewOtpErrorHandler(email).HandleError(err)

		if response != nil {
			return c.Status(fiber.StatusOK).JSON(response)
		}

		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to create payout account",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Payout account created successfully",
		Data:    payoutAccount,
	})
}

// @Summary Get payout account
// @Description Retrieve the payout account for the authenticated user
// @Tags payout
// @Accept json
// @Produce json
// @Param currency query string false "Currency code" default(NGN)
// @Success 200 {object} common.Response
// @Failure 400 {object} common.Response
// @Security Bearer
// @Router /payout/account [get]
func (s *PayoutHandler) GetPayoutAccount(c *fiber.Ctx) error {

	currency := c.Query("currency", "NGN")

	// currency = "USD"

	userId := auth.GetUserId(c)

	account, err := s.getPayoutAccount(currency, userId)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.Response{
			Status:  "error",
			Message: "Failed to fetch payout account",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(common.Response{
		Status:  "success",
		Message: "Payout account fetched successfully",
		Data:    account,
	})
}

// @Summary Get payout summary
// @Description Retrieve payout summary for a specific form
// @Tags payout
// @Accept json
// @Produce json
// @Param formId path string true "Form ID (UUID)"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /payout/summary/{formId} [get]
func (s *PayoutHandler) GetPayoutSummary(c *fiber.Ctx) error {

	userId := auth.GetUserId(c)
	formId, _ := uuid.Parse(c.Params("formId"))

	summary, err := s.getPayoutSummary(formId, userId)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to fetch payout summary",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Payout summary fetched successfully",
		Data:    summary,
	})
}

// @Summary Get payout history
// @Description Retrieve paginated payout history for a specific form
// @Tags payout
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
// @Router /payout/history/{formId} [get]
func (s *PayoutHandler) GetPayoutHistory(c *fiber.Ctx) error {

	userId := auth.GetUserId(c)
	formId, _ := uuid.Parse(c.Params("formId"))

	queryStr := string(c.Request().URI().QueryString())

	parsedModel := pager.ParseQueryString(queryStr)

	queryModel := PayoutQueryModel(parsedModel)

	payouts, err := s.getPayoutHistory(queryModel, formId, userId)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to fetch payout history",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Payout history fetched successfully",
		Data:    payouts,
	})
}

// @Summary Change payout type
// @Description Change the payout type for a payout account
// @Tags payout
// @Accept json
// @Produce json
// @Param request body ChangePayoutTypeRequest true "Payout type change request"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /payout/change-type [post]
func (s *PayoutHandler) ChangePayoutType(c *fiber.Ctx) error {

	var data ChangePayoutTypeRequest

	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	data.UserId = auth.GetUserId(c)

	payoutType, err := s.changePayoutType(data)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to change payout type",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Payout type changed successfully",
		Data:    payoutType,
	})
}

// @Summary Create payment payout provider
// @Description Create or update payment provider integration for payouts
// @Tags payout
// @Accept json
// @Produce json
// @Param request body CreatePaymentPayoutRequest true "Payment payout provider details"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Security Bearer
// @Router /payout/payment-provider [post]
func (s *PayoutHandler) CreatePaymentPayout(c *fiber.Ctx) error {

	var data CreatePaymentPayoutRequest

	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	userId := auth.GetUserId(c)

	data.UserId = userId

	err := s.createPaymentPayout(data)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Status:  "error",
			Message: "Failed to save payout provider integration",
			Error:   err.Error(),
		})
	}

	message := "Payout provider integration initiated successfully"
	if data.State == "completed" {
		message = "Payout provider integration completed successfully"
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: message,
		Data:    true,
	})
}

// @Summary Get payment payout provider
// @Description Retrieve payment provider integration details for the authenticated user
// @Tags payout
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Security Bearer
// @Router /payout/payment-provider [get]
func (s *PayoutHandler) GetPaymentPayout(c *fiber.Ctx) error {

	userId := auth.GetUserId(c)

	paymentPayout := s.getPaymentPayout(userId)

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Payment payout fetched successfully",
		Data:    paymentPayout,
	})
}

// @Summary Get total payout
// @Description Retrieve total payout amount for a specific form
// @Tags payout
// @Accept json
// @Produce json
// @Param formId path string true "Form ID (UUID)"
// @Success 200 {object} Response
// @Security Bearer
// @Router /payout/total-payout/{formId} [get]
func (s *PayoutHandler) GetTotalPayout(c *fiber.Ctx) error {

	formId := uuid.MustParse(c.Params("formId"))

	userId := auth.GetUserId(c)
	totalPayout := s.getTotalPayout(userId, formId)

	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Message: "Total payout fetched successfully",
		Data:    totalPayout,
	})
}
