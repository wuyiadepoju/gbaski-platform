package application

import (
	"context"
	"fmt"

	"github.com/gbaski/gbaski-ext/log"
	"github.com/gbaski/gbaski-platform/internal/payout/contracts"
	"github.com/gbaski/gbaski-platform/internal/payout/domain"
)

// CreatePaymentProviderService handles creating/updating a payment provider integration
type CreatePaymentProviderService struct {
	repo        contracts.PayoutRepository
	mailService contracts.MailService
	eventBus    contracts.EventBus
}

func NewCreatePaymentProviderService(
	repo contracts.PayoutRepository,
	mailService contracts.MailService,
	eventBus contracts.EventBus,
) *CreatePaymentProviderService {
	return &CreatePaymentProviderService{
		repo:        repo,
		mailService: mailService,
		eventBus:    eventBus,
	}
}

// Execute handles both 'initiated' (cache keys) and 'completed' (persist) states
func (s *CreatePaymentProviderService) Execute(ctx context.Context, cmd CreatePaymentProviderCommand) error {
	state, err := domain.NewProviderState(cmd.State)
	if err != nil {
		return fmt.Errorf("invalid provider state: %w", err)
	}

	if state.IsInitiated() {
		// Initiated state is handled by the handler layer via Redis (no domain logic needed)
		return nil
	}

	if state.IsCompleted() {
		provider, err := domain.NewPaymentProvider(cmd.UserID, cmd.Provider, cmd.PublicKey, cmd.SecretKey)
		if err != nil {
			return err
		}

		provider.CompleteIntegration()

		if err := s.repo.SavePaymentProvider(ctx, provider); err != nil {
			return fmt.Errorf("failed to save payment provider: %w", err)
		}

		// Publish domain events
		s.publishDomainEvents(provider)

		// Send email asynchronously
		go func() {
			if err := s.mailService.SendPayoutProviderLinkedEmail(cmd.UserID, cmd.Provider); err != nil {
				log.Error("payout", "send_provider_linked_email", err)
			}
		}()

		log.Info("payout", "create_payment_provider", map[string]interface{}{
			"user_id":  cmd.UserID,
			"provider": cmd.Provider,
		})
	}

	return nil
}

func (s *CreatePaymentProviderService) publishDomainEvents(provider *domain.PaymentProvider) {
	for _, event := range provider.DomainEvents() {
		s.eventBus.Publish(event)
	}
	provider.ClearDomainEvents()
}
