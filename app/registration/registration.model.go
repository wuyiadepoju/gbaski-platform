package registration

import (
	"encoding/json"
	"time"

	"github.com/gbaski/gbaski-platform/app/common"

	"github.com/google/uuid"
)

type Response = common.Response

type RegistrationStatus string

const (
	RegistrationStatusRegistered RegistrationStatus = "registered"
	RegistrationStatusCheckedIn  RegistrationStatus = "checkedin"
	RegistrationStatusCancelled  RegistrationStatus = "cancelled"
)

type RegistrationItem struct {
	Id           *int               `json:"id" db:"id"`
	FirstName    string             `json:"firstName" db:"first_name"`
	LastName     *string            `json:"lastName" db:"last_name"`
	Email        string             `json:"email" db:"email"`
	Phone        *string            `json:"phone" db:"phone"`
	Picture      *string            `json:"picture" db:"picture"`
	RegDesc      *string            `json:"regDesc" db:"reg_desc"`
	RegRef       string             `json:"regRef" db:"reg_ref"`
	RegStatus    RegistrationStatus `json:"regStatus" db:"reg_status"`
	CreatedAt    time.Time          `json:"date" db:"created_at"`
	Qty          *int               `json:"regQty" db:"qty"`
	Ref          *string            `json:"txRef" db:"ref"`
	TxId         *string            `json:"txId" db:"tx_id"`
	Amount       *float64           `json:"txAmount" db:"amount"`
	TxPro        *string            `json:"txProvider" db:"tx_pro"`
	TxCur        *string            `json:"txCurrency" db:"tx_cur"`
	TxDate       *time.Time         `json:"txDate" db:"tx_date"`
	Metadata     *json.RawMessage   `json:"metadata" db:"metadata"`
	SearchVector *string            `json:"-" db:"search_vector"`
}

type RegistrationList struct {
	Items []RegistrationItem `json:"items"`
	Page  int                `json:"page"`
	Size  int                `json:"size"`
	Total int                `json:"total"`
}

type RegistrationStats struct {
	ItemStats ItemStats `json:"itemStats"`
	SaleStats SaleStats `json:"saleStats"`
}

type ItemStats struct {
	TotalRegistered int `json:"totalRegistered"`
	TotalCheckedIn  int `json:"totalCheckedIn"`
	TotalCancelled  int `json:"totalCancelled"`
}

type SaleStats struct {
	TotalAmount   *float64 `json:"totalAmount"`
	TotalQty      *int     `json:"totalQty"`
	TotalRefunded *float64 `json:"totalRefunded"`
}

type RegistrationQueryModel struct {
	Search string              `json:"search"`
	Filter map[string]string   `json:"filter"`
	Range  map[string][]string `json:"range"`
	Page   int                 `json:"page"`
	Size   int                 `json:"size"`
}

type TicketAttendee struct {
	Id           int64      `json:"id" db:"id"`
	FirstName    string     `json:"firstName" db:"first_name"`
	LastName     *string    `json:"lastName" db:"last_name"`
	Email        string     `json:"email" db:"email"`
	Phone        string     `json:"phone" db:"phone"`
	RegRef       string     `json:"regRef" db:"reg_ref"`
	RegStatus    string     `json:"regStatus" db:"reg_status"`
	TicketId     int64      `json:"ticketId" db:"ticket_id"`
	TicketName   string     `json:"ticketName" db:"ticket_name"`
	CheckedIn    bool       `json:"checkedIn" db:"checked_in"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
	CheckedAt    *time.Time `json:"checkedAt" db:"checked_at"`
	SearchVector *string    `json:"-" db:"search_vector"`
}

type TicketAttendeeList struct {
	Items []TicketAttendee `json:"items"`
	Page  int              `json:"page"`
	Size  int              `json:"size"`
	Total int              `json:"total"`
}

type CheckInAgent struct {
	Name         string    `json:"name"`
	Id           uuid.UUID `json:"id"`
	TotalChecked int       `json:"totalChecked"`
	FormId       uuid.UUID `json:"formId"`
	RegType      string    `json:"regType"`
}
