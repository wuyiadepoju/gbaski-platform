# DDD Refactoring Summary for Event Module

## Overview
The event module has been refactored to follow Domain-Driven Design (DDD) principles while maintaining the existing layered architecture.

## New Structure

```
app/event/
├── domain/                    # Domain Layer (Business Logic)
│   ├── value_objects.go      # Value objects (EventName, EventDateRange, etc.)
│   ├── event.go              # Event aggregate root with business logic
│   ├── events.go             # Domain events
│   └── repository.go         # Repository interfaces
│
├── application/               # Application Layer (Use Cases)
│   ├── commands.go           # Command DTOs
│   ├── queries.go            # Query DTOs
│   ├── dtos.go               # Data Transfer Objects
│   ├── create_event_service.go
│   ├── get_event_service.go
│   ├── get_events_service.go
│   ├── update_event_service.go
│   ├── update_event_status_service.go
│   ├── delete_event_service.go
│   ├── update_event_image_service.go
│   └── update_event_video_service.go
│
├── infrastructure/            # Infrastructure Layer (Technical Concerns)
│   ├── event_repository.go   # Repository implementation
│   ├── category_repository.go
│   └── adapters.go           # External service adapters
│
└── (existing files)          # Legacy files (to be gradually replaced)
```

## Key DDD Concepts Implemented

### 1. Domain Layer
- **Value Objects**: Type-safe, immutable objects (EventName, EventDateRange, Location, etc.)
- **Aggregate Root**: Event entity with encapsulated business logic
- **Domain Events**: EventCreatedEvent, EventPublishedEvent, EventUpdatedEvent, etc.
- **Repository Interfaces**: Defined in domain layer, implemented in infrastructure

### 2. Application Layer
- **Commands**: CreateEventCommand, UpdateEventCommand, etc.
- **Queries**: GetEventQuery, GetEventsQuery, etc.
- **Application Services**: Orchestrate use cases, coordinate domain objects
- **DTOs**: Data transfer objects for API responses

### 3. Infrastructure Layer
- **Repository Implementations**: Bridge between domain and database
- **Adapters**: External service integrations (EventBus, WebmasterService, EmailScheduler)

## Migration Path

### Phase 1: ✅ Completed
- Created domain layer with value objects and entities
- Created application layer with services
- Created infrastructure layer foundations

### Phase 2: In Progress
- Update handlers to use application services
- Gradually replace legacy service methods

### Phase 3: Future
- Complete email scheduling migration
- Add domain event handlers
- Implement event sourcing (optional)

## Usage Example

### Old Way (Anemic Domain Model)
```go
// Business logic in service
func (s *Service) createUpdateEvent(event EventModel, user User) {
    status := EventStatusDraft
    if event.AccessType == AccessTypePublic && event.Payment == EventPaymentFree {
        status = EventStatusPublished
    }
    // ... more logic
}
```

### New Way (Rich Domain Model)
```go
// Business logic in domain entity
event, _ := domain.NewEvent(userId, name, description, dateRange, ...)
if accessType.IsPublic() && payment.IsFree() {
    event.Publish() // Encapsulated business rule
}

// Application service orchestrates
service := application.NewCreateEventService(...)
result, err := service.Execute(ctx, cmd)
```

## Benefits

1. **Business Logic in Domain**: Rules are encapsulated in entities
2. **Type Safety**: Value objects prevent invalid states
3. **Testability**: Domain logic can be tested without infrastructure
4. **Maintainability**: Clear separation of concerns
5. **Extensibility**: Easy to add new business rules

## Next Steps

1. Update handlers to use application services (see example below)
2. Migrate remaining service methods
3. Add comprehensive tests for domain logic
4. Implement domain event handlers
5. Consider event sourcing for audit trail

## Handler Migration Example

```go
// New DDD-compliant handler
type EventHandler struct {
    createService *application.CreateEventService
    getService    *application.GetEventService
    // ... other services
}

func (h *EventHandler) CreateEvent(c *fiber.Ctx) error {
    var req CreateEventRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(Response{...})
    }

    user := auth.GetUser(c)
    cmd := application.CreateEventCommand{
        UserID: user.Id,
        Name: req.Name,
        // ... map request to command
    }

    result, err := h.createService.Execute(c.Context(), cmd)
    if err != nil {
        return c.Status(400).JSON(Response{...})
    }

    return c.Status(201).JSON(Response{Data: result})
}
```

## Notes

- The refactoring maintains backward compatibility where possible
- Legacy code can coexist with new DDD code during migration
- All existing functionality is preserved
- New features should use the DDD structure
