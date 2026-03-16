package application

import (
	"github.com/gbaski/gbaski-platform/internal/payout/domain"
	"github.com/google/uuid"
)

// CreatePayoutAccountRequest is the HTTP request body for creating a payout account
type CreatePayoutAccountRequest struct {
	AccountNumber string    `json:"accountNumber"`
	AccountName   string    `json:"accountName"`
	BankID        int       `json:"bankId"`
	Currency      string    `json:"currency"`
	UserID        uuid.UUID `json:"userId"`
}

// ChangePayoutTypeRequest is the HTTP request body for changing payout type
type ChangePayoutTypeRequest struct {
	PayoutAccountID int       `json:"payoutAccountId"`
	PayoutType      string    `json:"payoutType"`
	UserID          uuid.UUID `json:"userId"`
}

// CreatePaymentProviderRequest is the HTTP request body for creating a payment provider integration
type CreatePaymentProviderRequest struct {
	UserID     uuid.UUID `json:"userId"`
	SellerName string    `json:"sellerName"`
	Provider   string    `json:"provider"`
	PublicKey  string    `json:"publicKey"`
	SecretKey  string    `json:"secretKey"`
	State      string    `json:"state"`
}

// PayoutAccountDTO is the response DTO for a payout account
type PayoutAccountDTO struct {
	ID            int         `json:"id"`
	AccountNumber string      `json:"accountNumber"`
	AccountName   string      `json:"accountName"`
	BankName      string      `json:"bankName"`
	BankID        int         `json:"bankId"`
	PayoutType    string      `json:"payoutType"`
	FeeRates      FeeRatesDTO `json:"feeRates"`
	Currency      string      `json:"currency"`
}

// FeeRatesDTO is the response DTO for fee rates
type FeeRatesDTO struct {
	Instant int `json:"instant"`
	Weekly  int `json:"weekly"`
}

// BankDTO is the response DTO for a bank
type BankDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// PaymentProviderDTO is the response DTO for a payment provider.
// Balance is stored as any since it holds a *wallet.WalletBalance from enrichment.
type PaymentProviderDTO struct {
	UserID    uuid.UUID `json:"userId"`
	Provider  string    `json:"provider"`
	PublicKey string    `json:"publicKey"`
	SecretKey string    `json:"secretKey,omitempty"`
	Balance   any       `json:"balance"`
}

// ToPayoutAccountDTO maps a domain PayoutAccount to its DTO
func ToPayoutAccountDTO(account *domain.PayoutAccount) PayoutAccountDTO {
	return PayoutAccountDTO{
		ID:            account.ID(),
		AccountNumber: account.AccountNumber(),
		AccountName:   account.AccountName(),
		BankName:      account.BankName(),
		BankID:        account.BankID(),
		PayoutType:    account.PayoutType().Value(),
		FeeRates: FeeRatesDTO{
			Instant: account.FeeRates().InstantRate(),
			Weekly:  account.FeeRates().WeeklyRate(),
		},
		Currency: account.Currency().Value(),
	}
}

// ToPaymentProviderDTO maps a domain PaymentProvider to its DTO
func ToPaymentProviderDTO(provider *domain.PaymentProvider) PaymentProviderDTO {
	return PaymentProviderDTO{
		UserID:    provider.UserID(),
		Provider:  provider.Provider(),
		PublicKey: provider.PublicKey(),
		SecretKey: provider.SecretKey(),
		Balance:   provider.Balance(),
	}
}

// ToBankDTO maps a domain Bank to its DTO
func ToBankDTO(bank domain.Bank) BankDTO {
	return BankDTO{
		ID:   bank.ID(),
		Name: bank.Name(),
		Code: bank.Code(),
	}
}
