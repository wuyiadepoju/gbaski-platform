package contracts

import (
	"context"
	"time"

	"github.com/gbaski/gbaski-platform/internal/payout/domain"
	"github.com/google/uuid"
)

// QueryParams represents query parameters for paginated payout history
type QueryParams struct {
	Search string
	Filter map[string]string
	Range  map[string][]string
	Page   int
	Size   int
}

// PayoutItem represents a single historical payout entry
type PayoutItem struct {
	Amount float32   `json:"amount" db:"amount"`
	TxRef  string    `json:"txRef" db:"tx_ref"`
	PaidAt time.Time `json:"paidAt" db:"paid_at"`
}

// PayoutList represents a paginated list of payout history items
type PayoutList struct {
	Items []PayoutItem `json:"items"`
	Page  int          `json:"page"`
	Size  int          `json:"size"`
	Total int          `json:"total"`
}

// PayoutSummaryDTO represents a payout summary for a given payout type
type PayoutSummaryDTO struct {
	PayoutType       domain.PayoutType `json:"payoutType"`
	NextPayoutDate   *time.Time        `json:"nextPayoutDate"`
	PendingAmount    float32           `json:"pendingAmount"`
	ProcessingAmount float32           `json:"processingAmount"`
}

// PayoutRepository defines the contract for payout persistence operations
type PayoutRepository interface {
	// Bank operations
	GetBanks(ctx context.Context) ([]domain.Bank, error)

	// PayoutAccount operations
	SavePayoutAccount(ctx context.Context, account *domain.PayoutAccount) error
	FindPayoutAccount(ctx context.Context, currency string, userID uuid.UUID) (*domain.PayoutAccount, error)
	ChangePayoutType(ctx context.Context, accountID int, userID uuid.UUID, payoutType domain.PayoutType) (*domain.PayoutAccount, error)

	// Payout ledger / history operations
	GetPayoutSummary(ctx context.Context, formID, userID uuid.UUID) (map[string]PayoutSummaryDTO, error)
	GetPayoutHistory(ctx context.Context, params QueryParams, formID, userID uuid.UUID) (PayoutList, error)
	GetTotalPayout(ctx context.Context, userID, formID uuid.UUID) float64

	// Payment provider operations
	SavePaymentProvider(ctx context.Context, provider *domain.PaymentProvider) error
	FindPaymentProviderBySellerName(ctx context.Context, sellerName string) (*domain.PaymentProvider, error)
	FindPaymentProviderByUserID(ctx context.Context, userID uuid.UUID) (*domain.PaymentProvider, error)
}
