package application

// GetEventLocation determines the event location based on the event's mode type and location.
// Returns "Online Event" as default, or the actual location if provided.
func GetEventLocation(event CreateEventRequest) string {
	eventLocation := "Online Event"
	if event.Location != nil && *event.Location != "" {
		eventLocation = *event.Location
	}
	if event.ModeType == "physical" && event.Location != nil {
		eventLocation = *event.Location
	}
	return eventLocation
}
