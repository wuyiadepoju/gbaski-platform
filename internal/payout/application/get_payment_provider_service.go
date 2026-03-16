package application

import (
	"context"

	"github.com/gbaski/gbaski-event/pkg/wallet"
	"github.com/gbaski/gbaski-platform/internal/payout/contracts"
)

// GetPaymentProviderService fetches the payment provider for a user, enriched with balance
type GetPaymentProviderService struct {
	repo contracts.PayoutRepository
}

func NewGetPaymentProviderService(repo contracts.PayoutRepository) *GetPaymentProviderService {
	return &GetPaymentProviderService{repo: repo}
}

// Execute returns the payment provider DTO with secret key redacted and balance enriched
func (s *GetPaymentProviderService) Execute(ctx context.Context, query GetPaymentProviderQuery) *PaymentProviderDTO {
	provider, err := s.repo.FindPaymentProviderByUserID(ctx, query.UserID)
	if err != nil || provider == nil {
		return nil
	}

	// Business rule: secret key must not be exposed on read
	provider.RedactSecretKey()

	// Enrich with live wallet balance
	walletService := wallet.NewWalletService()
	provider.SetBalance(walletService.GetBalance(provider.UserID()))

	dto := ToPaymentProviderDTO(provider)
	return &dto
}

// GetPaymentProviderBySellerService fetches the payment provider by seller name (used by other services)
type GetPaymentProviderBySellerService struct {
	repo contracts.PayoutRepository
}

func NewGetPaymentProviderBySellerService(repo contracts.PayoutRepository) *GetPaymentProviderBySellerService {
	return &GetPaymentProviderBySellerService{repo: repo}
}

// Execute returns the payment provider DTO for a seller (with secret key intact for internal use)
func (s *GetPaymentProviderBySellerService) Execute(ctx context.Context, query GetPaymentProviderBySellerQuery) (*PaymentProviderDTO, error) {
	provider, err := s.repo.FindPaymentProviderBySellerName(ctx, query.SellerName)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return nil, nil
	}
	dto := ToPaymentProviderDTO(provider)
	return &dto, nil
}
