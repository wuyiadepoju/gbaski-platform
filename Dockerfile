# Build stage
FROM public.ecr.aws/docker/library/golang:1.24-alpine AS builder

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy all required modules maintaining directory structure
# This preserves the relative paths needed for replace directives
COPY gbaski-shared/ ./gbaski-shared/
COPY gbaski-ext/ ./gbaski-ext/
COPY gbaski-event/ ./gbaski-event/
COPY gbaski-platform/ ./gbaski-platform/

# Set working directory to gbaski-platform for the build
WORKDIR /app/gbaski-platform

# Download dependencies and tidy up (after copying source code so replace directives work)
RUN go mod download && go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o gbaski-platform .

# Final stage
FROM alpine:3.19

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create app user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/gbaski-platform/gbaski-platform .

# Change ownership to appuser
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose the port
EXPOSE 8004

# Run the application
CMD ["./gbaski-platform"]

