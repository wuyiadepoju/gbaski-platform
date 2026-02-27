package middleware

import (
	"strings"

	"github.com/gbaski/gbaski-ext/auth"
	"github.com/gbaski/gbaski-ext/log"
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/gbaski/gbaski-shared/config"
)

func ApiKey(c *fiber.Ctx) error {

	apiKey := c.Get("x-api-key")

	// fmt.Println("API_Key: ", apiKey)

	// fmt.Printf("Request Headers: %s", c.Request().Header.Header())

	keys := []string{"cfb51695-bbix-3542-z5xb-e1244a71ffbe", "398f735d-eoxu-2238-e6ae-5e3531ae9f1a"}
	for _, key := range keys {
		if apiKey == key {
			return c.Next()
		}

	}

	return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized: Invalid API Key")
}

func Protected(c *fiber.Ctx) error {

	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(config.AppConfig.JWT_SECRET)},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status": "error",
				"error":  "Unauthorized: Invalid or expired token",
			})
		},
	})(c)

}

func AccessControl(c *fiber.Ctx) error {

	// Initialize Casbin enforcer
	// global.Enforcer, _ = casbin.NewEnforcer("./security/model.conf", "./security/policy.csv")
	// global.Enforcer.LoadPolicy()

	// e := global.Enforcer

	// session := session.NewUserContext()
	// session.SetUserContext(c)

	// user := session.GetUser()

	// sub := user.AccountType
	// obj := strings.TrimPrefix(c.Path(), "/api")
	// act := c.Method()

	// ok, err := e.Enforce(sub, obj, act)

	// if err != nil {
	// 	log.Printf("Error: %s", err)
	// }

	// if ok {
	// 	return c.Next()
	// }

	// log.Printf("Access denied")

	// return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
	// 	"status":  "error",
	// 	"message": "Unauthorized access",
	// })

	return c.Next()

}

// LogRequestContext logs request context information to Loki
// This helps identify which user and route made the request
func LogRequestContext(ctx *fiber.Ctx) error {
	requestData := make(map[string]interface{})

	// Add route information
	requestData["route"] = ctx.Route().Path
	requestData["method"] = ctx.Method()
	requestData["path"] = ctx.Path()
	requestData["ip"] = ctx.IP()

	// Add user context if authenticated
	if user := ctx.Locals("user"); user != nil {
		authUser := auth.GetUser(ctx)
		requestData["user_id"] = authUser.Id.String()
		requestData["user_email"] = authUser.Email
		requestData["user_name"] = authUser.Name
		requestData["account_type"] = string(authUser.AccountType)
	}

	// Add event/form IDs from URL params if present
	if eventId := ctx.Params("id"); eventId != "" {
		requestData["event_id"] = eventId
	}
	if formId := ctx.Params("formId"); formId != "" {
		requestData["form_id"] = formId
	}

	// Log request context
	log.Info("host", "request_context", requestData)

	return ctx.Next()
}

// ErrorHandler logs request errors to Loki
func ErrorHandler(ctx *fiber.Ctx) error {
	err := ctx.Next()
	if err != nil {
		errorData := map[string]interface{}{
			"route":  ctx.Route().Path,
			"method": ctx.Method(),
			"path":   ctx.Path(),
			"ip":     ctx.IP(),
			"status": ctx.Response().StatusCode(),
		}

		// Add user context if authenticated
		if user := ctx.Locals("user"); user != nil {
			authUser := auth.GetUser(ctx)
			errorData["user_id"] = authUser.Id.String()
			errorData["user_email"] = authUser.Email
		}

		log.Error("host", "request_error", err)
		log.Info("host", "error_context", errorData)
	}
	return err
}

// RecoverConfig returns a recover middleware configuration that logs panics to Loki
func RecoverConfig() recover.Config {
	return recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			errorData := map[string]interface{}{
				"route":  c.Route().Path,
				"method": c.Method(),
				"path":   c.Path(),
				"ip":     c.IP(),
				"panic":  e,
			}

			// Add user context if authenticated
			if user := c.Locals("user"); user != nil {
				authUser := auth.GetUser(c)
				errorData["user_id"] = authUser.Id.String()
				errorData["user_email"] = authUser.Email
			}

			log.Error("host", "panic_recovered", fiber.NewError(fiber.StatusInternalServerError, "panic recovered"))
			log.Info("host", "panic_context", errorData)
		},
	}
}

// SecurityMonitor monitors and logs suspicious query parameters
func SecurityMonitor(ctx *fiber.Ctx) error {
	queryString := string(ctx.Request().URI().QueryString())
	suspiciousPatterns := []string{
		"<script", "javascript:", "onerror=", "onload=",
		"union select", "drop table", "exec(", "eval(",
		"../", "..\\", "/etc/passwd", "boot.ini",
	}

	isSuspicious := false
	lowerQuery := strings.ToLower(queryString)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerQuery, pattern) {
			isSuspicious = true
			break
		}
	}

	if isSuspicious && queryString != "" {
		securityData := map[string]interface{}{
			"security_event": "suspicious_query",
			"severity":       "medium",
			"service":        "gbaski-host",
			"query_string":   queryString,
			"path":           ctx.Path(),
			"user_agent":     ctx.Get("User-Agent"),
			"ip":             ctx.IP(),
			"method":         ctx.Method(),
		}

		log.Warn("host", "suspicious_query_detected", securityData)
	}

	return ctx.Next()
}
