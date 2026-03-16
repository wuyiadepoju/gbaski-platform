package adapters

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/gbaski/gbaski-ext/log"
	"github.com/gbaski/gbaski-ext/mailcoach"
	"github.com/gbaski/gbaski-platform/internal/payout/contracts"
	"github.com/google/uuid"
)

// MailServiceImpl wraps mailcoach and the auth service to send payout emails
type MailServiceImpl struct {
	mailcoach   *mailcoach.Mailcoach
	authService contracts.AuthService
}

func NewMailServiceImpl(authService contracts.AuthService) contracts.MailService {
	return &MailServiceImpl{
		mailcoach:   mailcoach.New(),
		authService: authService,
	}
}

// SendPayoutProviderLinkedEmail sends an email confirming payment provider linking
func (m *MailServiceImpl) SendPayoutProviderLinkedEmail(userID uuid.UUID, provider string) error {
	user, err := m.authService.GetUserByID(userID)
	if err != nil {
		log.Error("payout", "get_user_by_id", err)
		return err
	}
	if user == nil || user.Email == "" {
		err := fmt.Errorf("user not found or email is empty for user_id: %s", userID)
		log.Error("payout", "send_payout_provider_linked_email", err)
		return err
	}

	type TemplateData struct {
		UserName string
		Provider string
	}

	htmlContent, err := renderTemplate("payout-provider-linked", TemplateData{
		UserName: user.Name,
		Provider: provider,
	})
	if err != nil {
		log.Error("payout", "render_template", err)
		return err
	}

	subject := "Welcome to Gbaski Pro - Payout Provider Linked"
	if err := m.mailcoach.SendSimpleTransactionalEmail(user.Email, subject, htmlContent); err != nil {
		log.Error("payout", "send_payout_provider_linked_email", err)
		return err
	}

	log.Info("payout", "send_payout_provider_linked_email", map[string]interface{}{
		"user_id":  userID,
		"email":    user.Email,
		"provider": provider,
	})
	return nil
}

func renderTemplate[T any](templateName string, data T) (string, error) {
	baseTemplatePath := filepath.Join("/var/www/gbaski/api/web/templates", "base.html")
	templatePath := filepath.Join("/var/www/gbaski/api/web/templates", fmt.Sprintf("%s.html", templateName))

	if _, err := os.Stat(baseTemplatePath); os.IsNotExist(err) {
		return "", fmt.Errorf("base template file not found: %s", baseTemplatePath)
	}
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return "", fmt.Errorf("template file not found: %s", templatePath)
	}

	tmpl, err := template.ParseFiles(baseTemplatePath, templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to parse templates: %w", err)
	}

	var htmlBuffer bytes.Buffer
	if err := tmpl.Execute(&htmlBuffer, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return htmlBuffer.String(), nil
}
