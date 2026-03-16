package repo

import (
	"github.com/gbaski/gbaski-platform/internal/event/domain"
	"github.com/google/uuid"
)

// SchemaTicketItemDTO represents a ticket item in the event schema
// This is a DTO used for database persistence and should not be in the domain layer
type SchemaTicketItemDTO struct {
	Id          int64                 `json:"id" db:"id"`
	UUID        *uuid.UUID            `json:"uuid" db:"uuid"`
	Name        string                `json:"name" db:"name"`
	Description *string               `json:"description,omitempty" db:"description"`
	Payment     domain.TicketPayment  `json:"payment" db:"payment"`
	Price       float64               `json:"price" db:"price"`
	Fee         float64               `json:"fee" db:"fee"`
	Discount    *int                  `json:"discount,omitempty" db:"discount"`
	Capacity    *int                  `json:"capacity,omitempty" db:"capacity"`
	Limit       *int                  `json:"limit,omitempty" db:"limit"`
	Sold        int                   `json:"sold" db:"sold"`
	Index       int                   `json:"index" db:"index"`
	Deleted     bool                  `json:"deleted" db:"deleted"`
}
