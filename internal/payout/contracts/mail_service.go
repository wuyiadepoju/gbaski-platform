package contracts

import "github.com/google/uuid"

// MailService defines the contract for sending payout-related emails
type MailService interface {
	SendPayoutProviderLinkedEmail(userID uuid.UUID, provider string) error
}
