package adapters

import (
	"context"
	"fmt"

	"github.com/gbaski/gbaski-platform/internal/event/contracts"
	"github.com/gbaski/gbaski-event/pkg/thirdparty/google"
)

// WebmasterServiceImpl implements the contracts.WebmasterService interface
type WebmasterServiceImpl struct{}

// NewWebmasterServiceImpl creates a new WebmasterServiceImpl
func NewWebmasterServiceImpl() contracts.WebmasterService {
	return &WebmasterServiceImpl{}
}

// AddEventURL adds an event URL to webmaster
func (s *WebmasterServiceImpl) AddEventURL(url string) error {
	webmasterService, err := google.NewWebmasterService(context.Background())
	if err != nil {
		return fmt.Errorf("failed to initialize webmaster service: %w", err)
	}

	if err := webmasterService.Add(url); err != nil {
		return fmt.Errorf("failed to add event URL: %w", err)
	}

	return nil
}
