package payout

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gbaski/gbaski-event/pkg/wallet"
	"github.com/gbaski/gbaski-ext/auth"
	"github.com/gbaski/gbaski-ext/log"
	"github.com/gbaski/gbaski-ext/mailcoach"
	"github.com/gbaski/gbaski-ext/otp"
	"github.com/gbaski/gbaski-shared/util"
	"github.com/google/uuid"
)

type Service struct {
	repo        *Repository
	mailcoach   *mailcoach.Mailcoach
	authHandler *auth.AuthHandler
}

func NewService() *Service {
	return &Service{
		repo:        NewRepository(),
		mailcoach:   mailcoach.New(),
		authHandler: auth.NewAuthHandler(),
	}
}

func (s *Service) getBanks() ([]Bank, error) {

	cacheKey := "payout_banks"

	cache := s.repo.Redis.Get(context.Background(), cacheKey)

	if cache.Val() != "" {
		var banks []Bank
		err := json.Unmarshal([]byte(cache.Val()), &banks)
		if err != nil {
			return nil, err
		}
		return banks, nil
	}

	banks, err := s.repo.GetBanks()

	if err != nil {
		return nil, err
	}

	jsonBanks, err := json.Marshal(banks)

	if err != nil {
		return nil, err
	}

	s.repo.Redis.Set(context.Background(), cacheKey, jsonBanks, 0)

	return banks, nil
}

func (s *Service) createPayoutAccount(data CreatePayoutAccountRequest, otpRequest otp.OtpRequest) (*PayoutAccount, error) {

	otpService := otp.NewOtpService[CreatePayoutAccountRequest, *PayoutAccount](
		otpRequest,
		otp.OtpServiceConfig{
			OtpData: data,
		},
	)

	otpResponse := otpService.ProcessOtp(context.Background())

	if !otpResponse.Success {
		return nil, otpResponse.Error
	}

	_, err := s.repo.CreatePayoutAccount(otpResponse.Data)
	if err != nil {
		return nil, err
	}

	payoutAccount, err := s.repo.GetPayoutAccount(otpResponse.Data.Currency, otpResponse.Data.UserId)
	if err != nil {
		return nil, err
	}

	return &PayoutAccount{
		AccountNumber: payoutAccount.AccountNumber,
		AccountName:   payoutAccount.AccountName,
		Currency:      payoutAccount.Currency,
		BankName:      payoutAccount.BankName,
		PayoutType:    payoutAccount.PayoutType,
		BankId:        payoutAccount.BankId,
		FeeRates:      payoutAccount.FeeRates,
	}, nil

}

func (s *Service) getPayoutAccount(currency string, userId uuid.UUID) (*PayoutAccount, error) {
	return s.repo.GetPayoutAccount(currency, userId)
}

func (s *Service) getPayoutSummary(userId uuid.UUID, formId uuid.UUID) (map[string]PayoutSummary, error) {
	return s.repo.GetPayoutSummary(userId, formId)
}

func (s *Service) getPayoutHistory(queryModel PayoutQueryModel, formId uuid.UUID, userId uuid.UUID) (PayoutList, error) {
	return s.repo.GetPayoutHistory(queryModel, formId, userId)
}

func (s *Service) changePayoutType(data ChangePayoutTypeRequest) (*PayoutAccount, error) {
	return s.repo.ChangePayoutType(data)
}

func (s *Service) createPaymentPayout(data CreatePaymentPayoutRequest) error {

	if data.State == "initiated" {

		key := fmt.Sprintf("payout_provider_integration:%s:%s", data.SellerName, strings.ToLower(string(data.Provider)))

		paymentPayout := PaymentPayout{
			UserId:    data.UserId,
			Provider:  data.Provider,
			PublicKey: data.PublicKey,
			SecretKey: data.SecretKey,
		}

		s.repo.Redis.Set(context.Background(), key, util.JsonStringify(paymentPayout), 1*time.Hour)
		return nil
	}

	if data.State == "completed" {
		err := s.repo.CreatePaymentPayout(data)
		if err != nil {
			return err
		}

		// Send email to acknowledge payout provider linking
		go s.sendPayoutProviderLinkedEmail(data.UserId, data.Provider)

		return nil
	}

	return nil
}

func (s *Service) GetPaymentPayoutBySellerName(sellerName string) (*PaymentPayout, error) {
	return s.repo.GetPaymentPayoutBySellerName(sellerName)
}

func (s *Service) GetPaymentPayoutByUserId(userId uuid.UUID) (*PaymentPayout, error) {
	return s.repo.GetPaymentPayoutByUserId(userId)
}

func (s *Service) getPaymentPayout(userId uuid.UUID) *PaymentPayout {

	paymentPayout, err := s.GetPaymentPayoutByUserId(userId)
	if err != nil {
		return nil
	}

	if paymentPayout == nil {
		return nil
	}

	paymentPayout.SecretKey = ""

	walletService := wallet.NewWalletService()
	paymentPayout.Balance = walletService.GetBalance(paymentPayout.UserId)

	return paymentPayout
}

func (s *Service) getTotalPayout(userId uuid.UUID, formId uuid.UUID) float64 {
	return s.repo.GetTotalPayout(userId, formId)
}

func (s *Service) sendPayoutProviderLinkedEmail(userId uuid.UUID, provider string) {
	user, err := s.authHandler.GetUserById(userId)
	if err != nil {
		log.Error("payout", "get_user_by_id", err)
		return
	}

	if user == nil || user.Email == "" {
		log.Error("payout", "send_payout_provider_linked_email", fmt.Errorf("user not found or email is empty for user_id: %s", userId))
		return
	}

	type PayoutProviderLinkedTemplateData struct {
		UserName string
		Provider string
	}

	templateData := PayoutProviderLinkedTemplateData{
		UserName: user.Name,
		Provider: provider,
	}

	htmlContent, err := renderTemplate("payout-provider-linked", templateData)
	if err != nil {
		log.Error("payout", "render_template", err)
		return
	}

	subject := "Welcome to Gbaski Pro - Payout Provider Linked"

	err = s.mailcoach.SendSimpleTransactionalEmail(user.Email, subject, htmlContent)
	if err != nil {
		log.Error("payout", "send_payout_provider_linked_email", err)
		return
	}

	log.Info("payout", "send_payout_provider_linked_email", map[string]interface{}{
		"user_id":  userId,
		"email":    user.Email,
		"provider": provider,
	})
}

func renderTemplate[T any](templateName string, data T) (string, error) {
	baseTemplatePath := filepath.Join("/var/www/gbaski/api/web/templates", "base.html")
	templatePath := filepath.Join("/var/www/gbaski/api/web/templates", fmt.Sprintf("%s.html", templateName))

	// Check if template files exist
	if _, err := os.Stat(baseTemplatePath); os.IsNotExist(err) {
		log.Error("payout", "find_base_template", fmt.Errorf("base template file not found: %s", baseTemplatePath))
		return "", fmt.Errorf("base template file not found: %s", baseTemplatePath)
	}

	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		log.Error("payout", "find_template", fmt.Errorf("template file not found: %s", templatePath))
		return "", fmt.Errorf("template file not found: %s", templatePath)
	}

	// Parse both the base template and the specific template
	tmpl, err := template.ParseFiles(baseTemplatePath, templatePath)
	if err != nil {
		log.Error("payout", "parse_templates", err)
		return "", fmt.Errorf("failed to parse templates: %w", err)
	}

	// Execute the template
	var htmlBuffer bytes.Buffer
	if err := tmpl.Execute(&htmlBuffer, data); err != nil {
		log.Error("payout", "execute_template", err)
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return htmlBuffer.String(), nil
}
