package event

import (
	"eventservice/src/internal/core"
	"fmt"
	"time"
)

type Service struct {
	repo core.EventRepository
}

func NewService(repo core.EventRepository) Service {
	return Service{repo: repo}
}

// CreateEvent creates a new event (for organizers)
func (s *Service) CreateEvent(req *core.CreateEventRequest, organizerID int) (*core.Event, error) {
	// Validate required fields
	if req.EventName == "" {
		return nil, fmt.Errorf("event name is required")
	}
	if req.Place == "" {
		return nil, fmt.Errorf("place is required")
	}
	if req.Capacity <= 0 {
		return nil, fmt.Errorf("capacity must be greater than 0")
	}

	// Parse and validate date
	eventDate, err := time.Parse("2006-01-02", req.EventDate)
	if err != nil {
		return nil, fmt.Errorf("invalid date format. Use YYYY-MM-DD")
	}

	// Validate time formats
	_, err = time.Parse("15:04", req.StartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start time format. Use HH:MM")
	}
	_, err = time.Parse("15:04", req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end time format. Use HH:MM")
	}

	// Check if the event date is in the future
	if eventDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return nil, fmt.Errorf("event date must be in the future")
	}

	// Create event object
	event := &core.Event{
		EventName:   req.EventName,
		OrganizerID: organizerID, // Use the organizerID from auth middleware
		Place:       req.Place,
		EventDate:   eventDate,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Capacity:    req.Capacity,
		Filled:      0,
		SeatsLeft:   req.Capacity,
	}

	return s.repo.CreateEvent(event)
}

// JoinEvent allows a customer to join an event
func (s *Service) JoinEvent(customerID int, eventID int) error {
	return s.repo.JoinEvent(customerID, eventID)
}

// GetAllEventsForCustomers gets all available events for customers with filters
func (s *Service) GetAllEventsForCustomers(filters *core.EventFilters) ([]core.EventResponse, error) {
	return s.repo.GetAllEventsForCustomers(filters)
}

// GetEventCustomers gets all customers who joined a specific event (for organizers)
func (s *Service) GetEventCustomers(eventID, organizerID int) ([]core.CustomerBooking, error) {
	return s.repo.GetEventCustomers(eventID, organizerID)
}

// GetOrganizerEvents gets all events created by an organizer
func (s *Service) GetOrganizerEvents(organizerID int) ([]core.Event, error) {
	return s.repo.GetOrganizerEvents(organizerID)
}

// GetEventByID gets a specific event by ID
func (s *Service) GetEventByID(eventID int) (*core.Event, error) {
	return s.repo.GetEventByID(eventID)
}

// GetAllEvents gets all events for customers with filters (renamed from GetAllEventsForCustomers)
func (s *Service) GetAllEvents(filters *core.EventFilters) ([]core.EventResponse, error) {
	return s.repo.GetAllEventsForCustomers(filters)
}

// GetEventsByOrganizer gets all events created by an organizer (alias for GetOrganizerEvents)
func (s *Service) GetEventsByOrganizer(organizerID int) ([]core.Event, error) {
	return s.repo.GetOrganizerEvents(organizerID)
}

// UpdateEvent updates an existing event
func (s *Service) UpdateEvent(eventID int, request *core.UpdateEventRequest, organizerID int) (*core.Event, error) {
	return s.repo.UpdateEvent(eventID, request, organizerID)
}

// DeleteEvent deletes an event
func (s *Service) DeleteEvent(eventID int, organizerID int) error {
	return s.repo.DeleteEvent(eventID, organizerID)
}

// JoinEventWithRequest allows a customer to join an event using a request object
func (s *Service) JoinEventWithRequest(userID int, request *core.JoinEventRequest) (*core.JoinEventResponse, error) {
	// For this implementation, we need to get user details from context or make them optional
	// Since we don't have access to email/username here, we'll use placeholder values
	err := s.repo.JoinEvent(userID, request.EventID)
	if err != nil {
		return nil, err
	}

	return &core.JoinEventResponse{
		Message: "Successfully joined event",
		EventID: request.EventID,
	}, nil
}

// LeaveEvent allows a customer to leave an event
func (s *Service) LeaveEvent(userID int, eventID int) error {
	return s.repo.LeaveEvent(userID, eventID)
}

// GetUserBookings gets all events a user has booked
func (s *Service) GetUserBookings(userID int) ([]core.Event, error) {
	return s.repo.GetUserBookings(userID)
}

// GetEventParticipants gets all participants for an event (alias for GetEventCustomers)
func (s *Service) GetEventParticipants(eventID, organizerID int) ([]core.CustomerBooking, error) {
	return s.repo.GetEventCustomers(eventID, organizerID)
}
