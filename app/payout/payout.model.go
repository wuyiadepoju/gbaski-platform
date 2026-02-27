package payout

import (
	"time"

	"github.com/gbaski/gbaski-event/pkg/event"
	"github.com/gbaski/gbaski-platform/app/common"

	"github.com/google/uuid"
)

type PayoutType string

const (
	PayoutTypeInstant PayoutType = "instant"
	PayoutTypeWeekly  PayoutType = "weekly"
)

type Response = common.Response

type PayoutAccount struct {
	Id            int    `json:"id" db:"id"`
	AccountNumber string `json:"accountNumber" db:"account_number"`
	AccountName   string `json:"accountName" db:"account_name"`
	BankName      string `json:"bankName" db:"bank_name"`
	BankId        int    `json:"bankId" db:"bank_id"`
	PayoutType    string `json:"payoutType" db:"payout_type"`
	FeeRates      struct {
		Instant int `json:"instant"`
		Weekly  int `json:"weekly"`
	} `json:"feeRates"`
	Currency string `json:"currency" db:"currency"`
}

type CreatePayoutAccountRequest struct {
	AccountNumber string    `json:"accountNumber"`
	AccountName   string    `json:"accountName"`
	BankId        int       `json:"bankId"`
	Currency      string    `json:"currency"`
	UserId        uuid.UUID `json:"userId"`
}

type Bank struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

type PayoutSummary struct {
	PayoutType       PayoutType `json:"payoutType"`
	NextPayoutDate   *time.Time `json:"nextPayoutDate"`
	PendingAmount    float32    `json:"pendingAmount"`
	ProcessingAmount float32    `json:"processingAmount"`
}

type PayoutItem struct {
	Amount float32   `json:"amount" db:"amount"`
	TxRef  string    `json:"txRef" db:"tx_ref"`
	PaidAt time.Time `json:"paidAt" db:"paid_at"`
}

type ChangePayoutTypeRequest struct {
	PayoutAccountId int       `json:"payoutAccountId"`
	PayoutType      string    `json:"payoutType"`
	UserId          uuid.UUID `json:"userId"`
}

type PayoutQueryModel struct {
	Search string              `json:"search"`
	Filter map[string]string   `json:"filter"`
	Range  map[string][]string `json:"range"`
	Page   int                 `json:"page"`
	Size   int                 `json:"size"`
}

type PayoutList struct {
	Items []PayoutItem `json:"items"`
	Page  int          `json:"page"`
	Size  int          `json:"size"`
	Total int          `json:"total"`
}

type SuspendPayout struct {
	Enabled bool      `json:"enabled"`
	Reason  string    `json:"reason"`
	Until   time.Time `json:"until"`
}

type PayoutConfig struct {
	InstantPayoutWeeklyLimit float64       `json:"instantPayoutWeeklyLimit"`
	SuspendPayout            SuspendPayout `json:"suspendPayout"`
}

type CreatePaymentPayoutRequest struct {
	UserId     uuid.UUID `json:"userId"`
	SellerName string    `json:"sellerName"`
	Provider   string    `json:"provider"`
	PublicKey  string    `json:"publicKey"`
	SecretKey  string    `json:"secretKey"`
	State      string    `json:"state"`
}

type PaymentPayout = event.PayoutProviderItem
