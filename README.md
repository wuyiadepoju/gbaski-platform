# Gbaski Host API (gbaski-platform)

The **Gbaski Host API** is the core backend service for the Gbaski platform, providing functionalities for event management, registration, payouts, and user settings. It is built with Go and the Fiber web framework, designed to run both as a standalone server and as an AWS Lambda function.

## 🚀 Features

- **Authentication**: JWT-based authentication, OTP login, and registration.
- **Event Management**: CRUD operations for events, categories, and statistics.
- **Registrations**: Ticket-based registrations, attendee management, and check-in agent tools.
- **Payouts**: Bank account management, payout history, and automated payout summaries.
- **Wallet**: Balance tracking and transaction management.
- **File Uploads**: S3 pre-signed URL generation for media uploads.
- **Developer Tools**: Integrated Swagger UI for API documentation and exploration.
- **Observability**: Structured logging with Loki and error tracking with Sentry.

## 🛠 Tech Stack

- **Language**: [Go](https://golang.org/) (1.24.1)
- **Web Framework**: [Fiber v2](https://gofiber.io/)
- **Database**: PostgreSQL (via [sqlx](https://github.com/jmoiron/sqlx))
- **Cache**: Redis
- **Cloud**: AWS (Lambda, S3, SQS, CloudFront)
- **Monitoring**: Loki, Sentry
- **Documentation**: [Swagger / Swag](https://github.com/swaggo/swag)

## 📁 Project Structure

```text
gbaski-platform/
├── cmd/                # Custom commands (e.g., database seeding)
├── docs/               # Generated Swagger documentation
├── internal/           # Private application code
│   ├── common/         # Shared utilities
│   ├── event/          # Event management logic
│   ├── form/           # Registration forms
│   ├── payout/         # Payout and banking logic
│   ├── registration/   # Attendee and ticket registration
│   └── setting/        # User and system settings
├── middleware/         # Fiber middlewares
├── scripts/            # Maintenance and automation scripts
├── main.go             # Application entry point
├── routes.go           # Route definitions
├── Dockerfile          # Container configuration
└── lambda.handler.go   # AWS Lambda shim
```

## 🚥 Getting Started

### Prerequisites

- Go 1.24+
- PostgreSQL
- [Air](https://github.com/cosmtrek/air) (for live reloading)

### Installation

1. Clone the repository and navigate to the project directory:
   ```bash
   cd gbaski-platform
   ```

2. Download dependencies:
   ```bash
   go mod download
   ```

3. Configure environment variables (refer to `gbaski-shared/config` for required keys).

### Running Locally

To start the server with live reloading:
```bash
air
```

Or run directly:
```bash
go run .
```

The API will be available at `http://localhost:8004` (depending on your configuration).

### API Documentation

Access the Swagger UI at:
`http://localhost:8004/swagger/index.html`

## 📦 Deployment

### Docker

Build the Docker image:
```bash
docker build -t gbaski-platform -f Dockerfile ..
```

### AWS Lambda

The project includes a `lambda.handler.go` which allows the Fiber app to run as an AWS Lambda function using the `aws-lambda-go-api-proxy`.

## 📜 License

This project is licensed under the Apache 2.0 License.
