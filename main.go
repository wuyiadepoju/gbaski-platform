// @title Gbaski Host API
// @version 1.0
// @description API documentation for Gbaski Host service
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@gbaski.app

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http https

package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gbaski/gbaski-event/pkg/wallet"
	"github.com/gbaski/gbaski-platform/internal/event"
	"github.com/gbaski/gbaski-platform/internal/payout"
	"github.com/gbaski/gbaski-platform/internal/registration"
	"github.com/gbaski/gbaski-platform/internal/setting"
	"github.com/gbaski/gbaski-shared/config"
	postgres "github.com/gbaski/gbaski-shared/postgres"
	"github.com/jmoiron/sqlx"

	fiberadapter "github.com/awslabs/aws-lambda-go-api-proxy/fiber"
	"github.com/gofiber/fiber/v2"

	"github.com/gbaski/gbaski-ext/auth"
	"github.com/gbaski/gbaski-ext/job"
	"github.com/gbaski/gbaski-ext/log"
	"github.com/gbaski/gbaski-ext/mailcoach"
)

var fiberLambda *fiberadapter.FiberLambda
var app *fiber.App

var (
	jobService          *job.Service
	mailcoachService    *mailcoach.Mailcoach
	authHandler         *auth.AuthHandler
	eventHandler        *event.EventHandler
	payoutHandler       *payout.PayoutHandler
	registrationHandler *registration.RegistrationHandler
	settingHandler      *setting.SettingHandler
	walletHandler       *wallet.WalletHandler
	db                  *sqlx.DB
)

func init() {
	jobService = job.NewService()
	mailcoachService = mailcoach.New()
	authHandler = auth.NewAuthHandler()
	// Event handler will be initialized in main() after config is loaded
	// eventHandler = event.NewEventHandler(db)
	payoutHandler = payout.NewPayoutHandler()
	registrationHandler = registration.NewRegistrationHandler()
	settingHandler = setting.NewSettingHandler()
	walletHandler = wallet.NewWalletHandler()
}

func main() {
	// Load configuration
	if err := config.LoadConfig(); err != nil {
		log.Error("platform", "load_config_error", err)
	}

	// Initialize database connection
	db = postgres.New()

	// Initialize event handler with DB connection
	eventHandler = event.NewEventHandler(db)

	// Setup Loki logging
	log.SetLokiConfig(log.LokiConfig{
		Labels: map[string]string{
			"service":     "gbaski-platform",
			"environment": config.AppConfig.API_ENV,
		},
		BatchSize: 10,
		BatchWait: 5 * time.Second,
	})

	log.Info("platform", "api_started", map[string]interface{}{
		"service": "gbaski-platform",
		"port":    config.AppConfig.API_PLATFORM_PORT,
	})

	// Create new Fiber app
	app = fiber.New(fiber.Config{
		AppName: fmt.Sprintf("%s-%s", config.AppConfig.API_NAME, "platform"),
	})

	// Setup routes
	SetupRoutes(app)

	serverAddr := fmt.Sprintf(":%s", config.AppConfig.API_PLATFORM_PORT)

	if err := app.Listen(serverAddr); err != nil && err != http.ErrServerClosed {
		log.Error("platform", "fiber_server_failed_to_start", err)
	}

}
