package application

import (
	"github.com/gbaski/gbaski-platform/internal/payout/contracts"
	"github.com/google/uuid"
)

// CreatePayoutAccountCommand is the command to create or update a payout account
type CreatePayoutAccountCommand struct {
	UserID        uuid.UUID
	AccountNumber string
	AccountName   string
	BankID        int
	Currency      string // "NGN" | "USD"
}

// ChangePayoutTypeCommand is the command to change payout type on an account
type ChangePayoutTypeCommand struct {
	PayoutAccountID int
	UserID          uuid.UUID
	PayoutType      string // "instant" | "weekly"
}

// CreatePaymentProviderCommand is the command to create/update a payment provider integration
type CreatePaymentProviderCommand struct {
	UserID     uuid.UUID
	SellerName string
	Provider   string
	PublicKey  string
	SecretKey  string
	State      string // "initiated" | "completed"
}

// GetPayoutAccountQuery is the query to fetch a user's payout account
type GetPayoutAccountQuery struct {
	Currency string
	UserID   uuid.UUID
}

// GetPayoutSummaryQuery is the query to fetch the payout summary for a form
type GetPayoutSummaryQuery struct {
	FormID uuid.UUID
	UserID uuid.UUID
}

// GetPayoutHistoryQuery is the query to fetch paginated payout history
type GetPayoutHistoryQuery struct {
	Params contracts.QueryParams
	FormID uuid.UUID
	UserID uuid.UUID
}

// GetTotalPayoutQuery is the query to fetch the total payout for a form
type GetTotalPayoutQuery struct {
	UserID uuid.UUID
	FormID uuid.UUID
}

// GetPaymentProviderQuery is the query to fetch the current user's payment provider
type GetPaymentProviderQuery struct {
	UserID uuid.UUID
}

// GetPaymentProviderBySellerQuery is the query to fetch a payment provider by seller name
type GetPaymentProviderBySellerQuery struct {
	SellerName string
}
