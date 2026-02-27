package event

// GetEventLocation determines the event location based on the event's mode type and location.
// Returns "Online Event" as default, or the actual location if provided.
func GetEventLocation(event EventModel) string {
	eventLocation := "Online Event"
	if event.Location != nil && *event.Location != "" {
		eventLocation = *event.Location
	}
	if event.ModeType == EventModePhysical && event.Location != nil {
		eventLocation = *event.Location
	}
	return eventLocation
}
