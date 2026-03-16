package application

import (
	"context"
	"fmt"

	"github.com/gbaski/gbaski-ext/log"
	"github.com/gbaski/gbaski-ext/otp"
	"github.com/gbaski/gbaski-platform/internal/payout/contracts"
	"github.com/gbaski/gbaski-platform/internal/payout/domain"
)

// CreatePayoutAccountService handles creating or updating a payout account
type CreatePayoutAccountService struct {
	repo     contracts.PayoutRepository
	eventBus contracts.EventBus
}

func NewCreatePayoutAccountService(
	repo contracts.PayoutRepository,
	eventBus contracts.EventBus,
) *CreatePayoutAccountService {
	return &CreatePayoutAccountService{
		repo:     repo,
		eventBus: eventBus,
	}
}

// Execute processes OTP and creates the payout account if valid
func (s *CreatePayoutAccountService) Execute(ctx context.Context, cmd CreatePayoutAccountCommand, otpRequest otp.OtpRequest) (*PayoutAccountDTO, error) {
	otpService := otp.NewOtpService[CreatePayoutAccountCommand, *PayoutAccountDTO](
		otpRequest,
		otp.OtpServiceConfig{
			OtpData: cmd,
		},
	)

	otpResponse := otpService.ProcessOtp(ctx)
	if !otpResponse.Success {
		return nil, otpResponse.Error
	}

	currency, err := domain.NewCurrency(cmd.Currency)
	if err != nil {
		return nil, fmt.Errorf("invalid currency: %w", err)
	}

	account, err := domain.NewPayoutAccount(
		cmd.UserID,
		cmd.AccountNumber,
		cmd.AccountName,
		cmd.BankID,
		currency,
	)
	if err != nil {
		return nil, err
	}

	if err := s.repo.SavePayoutAccount(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to save payout account: %w", err)
	}

	// Publish domain events
	s.publishDomainEvents(account)

	// Fetch the persisted account to get bank name from joined data
	persisted, err := s.repo.FindPayoutAccount(ctx, cmd.Currency, cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created payout account: %w", err)
	}

	log.Info("payout", "create_payout_account", map[string]interface{}{
		"user_id":  cmd.UserID,
		"currency": cmd.Currency,
	})

	dto := ToPayoutAccountDTO(persisted)
	return &dto, nil
}

func (s *CreatePayoutAccountService) publishDomainEvents(account *domain.PayoutAccount) {
	for _, event := range account.DomainEvents() {
		s.eventBus.Publish(event)
	}
	account.ClearDomainEvents()
}
