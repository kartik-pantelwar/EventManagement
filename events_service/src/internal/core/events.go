package core

import "time"

// Event represents an event in the system
type Event struct {
	EventID      int       `json:"event_id"`
	EventName    string    `json:"event_name"`
	OrganizerID  int       `json:"organizer_id"`
	Place        string    `json:"place"`
	EventDate    time.Time `json:"-"`          // Internal use only
	EventDateStr string    `json:"event_date"` // For JSON output: YYYY-MM-DD
	StartTime    string    `json:"start_time"` // Format: HH:MM
	EndTime      string    `json:"end_time"`   // Format: HH:MM
	Capacity     int       `json:"capacity"`
	Filled       int       `json:"filled"`
	SeatsLeft    int       `json:"seats_left"` // Calculated field: capacity - filled
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateEventRequest represents the request to create an event
type CreateEventRequest struct {
	EventName   string `json:"event_name" validate:"required"`
	OrganizerID int    `json:"organizer_id"` // This will be set from session
	Place       string `json:"place" validate:"required"`
	EventDate   string `json:"event_date" validate:"required"` // Format: YYYY-MM-DD
	StartTime   string `json:"start_time" validate:"required"` // Format: HH:MM
	EndTime     string `json:"end_time" validate:"required"`   // Format: HH:MM
	Capacity    int    `json:"capacity" validate:"required,min=1"`
}

// EventResponse represents the response for customers viewing events
type EventResponse struct {
	EventID       int    `json:"id"`
	EventName     string `json:"name"`
	OrganizerID   int    `json:"organizer"`
	OrganizerName string `json:"organizer_name"`
	Place         string `json:"place"`
	EventDate     string `json:"date"`       // Format: YYYY-MM-DD
	StartTime     string `json:"start_time"` // Format: HH:MM
	EndTime       string `json:"end_time"`   // Format: HH:MM
	Capacity      int    `json:"capacity"`
	SeatsLeft     int    `json:"seats_left"`
}

// EventFilters represents filters for event listing
type EventFilters struct {
	Date        string `json:"date,omitempty"` // YYYY-MM-DD
	Place       string `json:"place,omitempty"`
	OrganizerID int    `json:"organizer,omitempty"`
}

// JoinEventRequest represents the request to join an event by event ID
type JoinEventRequest struct {
	EventID int `json:"event_id" validate:"required"`
}

// CustomerBooking represents a customer's booking information
type CustomerBooking struct {
	EventID   int       `json:"event_id"`
	CID       int       `json:"cid"`
	CEmail    string    `json:"cemail"`
	CUsername string    `json:"cusername"`
	BookedAt  time.Time `json:"booked_at"`
}

// JoinEventResponse represents the response when a customer joins an event
type JoinEventResponse struct {
	Message string `json:"message"`
	EventID int    `json:"event_id"`
}

// UpdateEventRequest represents the request to update an event
type UpdateEventRequest struct {
	EventName string `json:"event_name,omitempty"`
	Place     string `json:"place,omitempty"`
	EventDate string `json:"event_date,omitempty"` // Format: YYYY-MM-DD
	StartTime string `json:"start_time,omitempty"` // Format: HH:MM
	EndTime   string `json:"end_time,omitempty"`   // Format: HH:MM
	Capacity  int    `json:"capacity,omitempty"`
}

// EventRepository defines the interface for event data operations
type EventRepository interface {
	CreateEvent(event *Event) (*Event, error)
	GetEventByID(eventID int) (*Event, error)
	GetAllEventsForCustomers(filters *EventFilters) ([]EventResponse, error)
	JoinEvent(customerID int, eventID int) error
	LeaveEvent(customerID int, eventID int) error
	GetEventCustomers(eventID, organizerID int) ([]CustomerBooking, error)
	GetOrganizerEvents(organizerID int) ([]Event, error)
	GetUserBookings(userID int) ([]Event, error)
	UpdateEvent(eventID int, request *UpdateEventRequest, organizerID int) (*Event, error)
	DeleteEvent(eventID int, organizerID int) error
}
