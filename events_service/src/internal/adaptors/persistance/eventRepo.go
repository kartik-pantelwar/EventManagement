package persistance

import (
	"database/sql"
	"eventservice/src/internal/core"
	"fmt"
	"strings"
	"time"
)

type EventRepo struct {
	db *Database
}

func NewEventRepo(d *Database) EventRepo {
	return EventRepo{db: d}
}

// CreateEvent creates a new event (organizer functionality)
func (er *EventRepo) CreateEvent(event *core.Event) (*core.Event, error) {
	// Check place availability first
	var placeAvailable bool
	checkQuery := `SELECT events_schema.check_place_availability($1, $2, $3, $4)`
	err := er.db.db.QueryRow(checkQuery, event.Place, event.EventDate, event.StartTime, event.EndTime).Scan(&placeAvailable)
	if err != nil {
		return nil, fmt.Errorf("failed to check place availability: %v", err)
	}

	if !placeAvailable {
		return nil, fmt.Errorf("place '%s' is not available for the given time slot", event.Place)
	}

	// Create the event
	query := `
		INSERT INTO events_schema.events (event_name, organizer_id, place, event_date, start_time, end_time, capacity)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING event_id, created_at, updated_at`

	var createdEvent core.Event
	err = er.db.db.QueryRow(query, event.EventName, event.OrganizerID,
		event.Place, event.EventDate, event.StartTime, event.EndTime, event.Capacity).
		Scan(&createdEvent.EventID, &createdEvent.CreatedAt, &createdEvent.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create event: %v", err)
	}

	// Copy the input data to the created event
	createdEvent.EventName = event.EventName
	createdEvent.OrganizerID = event.OrganizerID
	createdEvent.Place = event.Place
	createdEvent.EventDate = event.EventDate
	createdEvent.StartTime = event.StartTime
	createdEvent.EndTime = event.EndTime
	createdEvent.Capacity = event.Capacity
	createdEvent.Filled = 0
	createdEvent.SeatsLeft = event.Capacity

	return &createdEvent, nil
}

// GetAllEventsForCustomers returns all events with organizer name (public endpoint)
func (er *EventRepo) GetAllEventsForCustomers(filters *core.EventFilters) ([]core.EventResponse, error) {
	query := `
		SELECT 
			e.event_id, e.event_name, e.organizer_id, e.place, 
			e.event_date, e.start_time, e.end_time, e.capacity, 
			e.filled, e.created_at, e.updated_at,
			u.username as organizer_name
		FROM events_schema.events e
		JOIN users u ON e.organizer_id = u.cid
	`

	var args []interface{}
	argIndex := 1
	var conditions []string

	// Apply filters
	if filters.Date != "" {
		conditions = append(conditions, fmt.Sprintf("e.event_date = $%d", argIndex))
		args = append(args, filters.Date)
		argIndex++
	}

	if filters.Place != "" {
		conditions = append(conditions, fmt.Sprintf("LOWER(e.place) LIKE LOWER($%d)", argIndex))
		args = append(args, "%"+filters.Place+"%")
		argIndex++
	}

	if filters.OrganizerID != 0 {
		conditions = append(conditions, fmt.Sprintf("e.organizer_id = $%d", argIndex))
		args = append(args, filters.OrganizerID)
		argIndex++
	}

	// Add WHERE clause if we have conditions
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY e.event_date ASC, e.start_time ASC"

	rows, err := er.db.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %v", err)
	}
	defer rows.Close()

	var events []core.EventResponse
	for rows.Next() {
		var event core.EventResponse
		var filled int
		var createdAt, updatedAt time.Time
		err := rows.Scan(
			&event.EventID, &event.EventName, &event.OrganizerID, &event.Place,
			&event.EventDate, &event.StartTime, &event.EndTime, &event.Capacity,
			&filled, &createdAt, &updatedAt, &event.OrganizerName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %v", err)
		}
		event.SeatsLeft = event.Capacity - filled
		events = append(events, event)
	}

	return events, nil
}

// JoinEvent allows a customer to join an event by event ID
func (er *EventRepo) JoinEvent(customerID int, eventID int, customerEmail, customerUsername string) error {
	// First get event details
	var eventDate time.Time
	var startTime, endTime string
	var capacity, filled int

	eventQuery := `SELECT event_date, start_time, end_time, capacity, filled FROM events_schema.events WHERE event_id = $1`
	err := er.db.db.QueryRow(eventQuery, eventID).Scan(&eventDate, &startTime, &endTime, &capacity, &filled)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("event not found")
		}
		return fmt.Errorf("failed to get event details: %v", err)
	}

	// Check if event is full
	if filled >= capacity {
		return fmt.Errorf("event is full")
	}

	// Check for customer time conflicts
	var hasConflict bool
	conflictQuery := `SELECT events_schema.check_customer_time_conflict($1, $2, $3, $4)`
	err = er.db.db.QueryRow(conflictQuery, customerID, eventDate, startTime, endTime).Scan(&hasConflict)
	if err != nil {
		return fmt.Errorf("failed to check time conflict: %v", err)
	}

	if hasConflict {
		return fmt.Errorf("you already have an event during this time period")
	}

	// Join the event (we no longer need cemail and cusername since we can get them from users table)
	insertQuery := `
		INSERT INTO events_schema.userbooked_events (event_id, cid)
		VALUES ($1, $2)`

	_, err = er.db.db.Exec(insertQuery, eventID, customerID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return fmt.Errorf("you have already joined this event")
		}
		return fmt.Errorf("failed to join event: %v", err)
	}

	return nil
}

// GetEventCustomers gets all customers who have booked a specific event (organizer functionality)
func (er *EventRepo) GetEventCustomers(eventID int, organizerID int) ([]core.CustomerBooking, error) {
	// First verify the organizer owns this event
	ownerQuery := `SELECT COUNT(*) FROM events_schema.events WHERE event_id = $1 AND organizer_id = $2`
	var count int
	err := er.db.db.QueryRow(ownerQuery, eventID, organizerID).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to verify event ownership: %v", err)
	}

	if count == 0 {
		return nil, fmt.Errorf("event not found or you don't have permission to view its customers")
	}

	// Get customers with user details from auth service
	query := `
		SELECT 
			ub.cid, u.username, u.email, ub.booked_at
		FROM events_schema.userbooked_events ub
		JOIN users u ON ub.cid = u.cid
		WHERE ub.event_id = $1
		ORDER BY ub.booked_at ASC
	`

	rows, err := er.db.db.Query(query, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event customers: %v", err)
	}
	defer rows.Close()

	var customers []core.CustomerBooking
	for rows.Next() {
		var customer core.CustomerBooking
		err := rows.Scan(&customer.CID, &customer.CUsername, &customer.CEmail, &customer.BookedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan customer: %v", err)
		}
		customer.EventID = eventID // Set the event ID
		customers = append(customers, customer)
	}

	return customers, nil
}

// GetOrganizerEvents retrieves all events created by an organizer
func (er *EventRepo) GetOrganizerEvents(organizerID int) ([]core.Event, error) {
	query := `
		SELECT event_id, event_name, organizer_id, place, event_date, start_time, end_time, 
			   capacity, filled, created_at, updated_at
		FROM events_schema.events WHERE organizer_id = $1 ORDER BY event_date, start_time`

	rows, err := er.db.db.Query(query, organizerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organizer events: %v", err)
	}
	defer rows.Close()

	var events []core.Event
	for rows.Next() {
		var event core.Event
		err := rows.Scan(
			&event.EventID, &event.EventName, &event.OrganizerID,
			&event.Place, &event.EventDate, &event.StartTime, &event.EndTime,
			&event.Capacity, &event.Filled, &event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %v", err)
		}
		event.SeatsLeft = event.Capacity - event.Filled
		events = append(events, event)
	}

	return events, nil
}

// GetEventByID retrieves a specific event by ID
func (er *EventRepo) GetEventByID(eventID int) (*core.Event, error) {
	query := `
		SELECT event_id, event_name, organizer_id, place, event_date, start_time, end_time, 
			   capacity, filled, created_at, updated_at
		FROM events_schema.events WHERE event_id = $1`

	var event core.Event
	err := er.db.db.QueryRow(query, eventID).Scan(
		&event.EventID, &event.EventName, &event.OrganizerID,
		&event.Place, &event.EventDate, &event.StartTime, &event.EndTime,
		&event.Capacity, &event.Filled, &event.CreatedAt, &event.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("event not found")
		}
		return nil, fmt.Errorf("failed to get event: %v", err)
	}

	event.SeatsLeft = event.Capacity - event.Filled
	return &event, nil
}

// LeaveEvent allows a customer to leave an event
func (er *EventRepo) LeaveEvent(customerID int, eventID int) error {
	// First check if the customer is actually booked for this event
	checkQuery := `SELECT COUNT(*) FROM events_schema.userbooked_events WHERE cid = $1 AND event_id = $2`
	var count int
	err := er.db.db.QueryRow(checkQuery, customerID, eventID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check booking: %v", err)
	}

	if count == 0 {
		return fmt.Errorf("you are not booked for this event")
	}

	// Remove the booking
	deleteQuery := `DELETE FROM events_schema.userbooked_events WHERE cid = $1 AND event_id = $2`
	_, err = er.db.db.Exec(deleteQuery, customerID, eventID)
	if err != nil {
		return fmt.Errorf("failed to leave event: %v", err)
	}

	return nil
}

// GetUserBookings retrieves all events a user has booked
func (er *EventRepo) GetUserBookings(userID int) ([]core.Event, error) {
	query := `
		SELECT e.event_id, e.event_name, e.organizer_id, e.place, e.event_date, 
			   e.start_time, e.end_time, e.capacity, e.filled, e.created_at, e.updated_at
		FROM events_schema.events e
		JOIN events_schema.userbooked_events ub ON e.event_id = ub.event_id
		WHERE ub.cid = $1
		ORDER BY e.event_date, e.start_time`

	rows, err := er.db.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user bookings: %v", err)
	}
	defer rows.Close()

	var events []core.Event
	for rows.Next() {
		var event core.Event
		err := rows.Scan(
			&event.EventID, &event.EventName, &event.OrganizerID,
			&event.Place, &event.EventDate, &event.StartTime, &event.EndTime,
			&event.Capacity, &event.Filled, &event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %v", err)
		}
		event.SeatsLeft = event.Capacity - event.Filled
		events = append(events, event)
	}

	return events, nil
}

// UpdateEvent updates an existing event (organizer functionality)
func (er *EventRepo) UpdateEvent(eventID int, request *core.UpdateEventRequest, organizerID int) (*core.Event, error) {
	// First verify the organizer owns this event
	ownerQuery := `SELECT COUNT(*) FROM events_schema.events WHERE event_id = $1 AND organizer_id = $2`
	var count int
	err := er.db.db.QueryRow(ownerQuery, eventID, organizerID).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to verify event ownership: %v", err)
	}

	if count == 0 {
		return nil, fmt.Errorf("event not found or you don't have permission to update it")
	}

	// Build dynamic update query
	var setParts []string
	var args []interface{}
	argIndex := 1

	if request.EventName != "" {
		setParts = append(setParts, fmt.Sprintf("event_name = $%d", argIndex))
		args = append(args, request.EventName)
		argIndex++
	}

	if request.Place != "" {
		setParts = append(setParts, fmt.Sprintf("place = $%d", argIndex))
		args = append(args, request.Place)
		argIndex++
	}

	if request.EventDate != "" {
		eventDate, err := time.Parse("2006-01-02", request.EventDate)
		if err != nil {
			return nil, fmt.Errorf("invalid date format. Use YYYY-MM-DD")
		}
		if eventDate.Before(time.Now().Truncate(24 * time.Hour)) {
			return nil, fmt.Errorf("event date must be in the future")
		}
		setParts = append(setParts, fmt.Sprintf("event_date = $%d", argIndex))
		args = append(args, eventDate)
		argIndex++
	}

	if request.StartTime != "" {
		_, err := time.Parse("15:04", request.StartTime)
		if err != nil {
			return nil, fmt.Errorf("invalid start time format. Use HH:MM")
		}
		setParts = append(setParts, fmt.Sprintf("start_time = $%d", argIndex))
		args = append(args, request.StartTime)
		argIndex++
	}

	if request.EndTime != "" {
		_, err := time.Parse("15:04", request.EndTime)
		if err != nil {
			return nil, fmt.Errorf("invalid end time format. Use HH:MM")
		}
		setParts = append(setParts, fmt.Sprintf("end_time = $%d", argIndex))
		args = append(args, request.EndTime)
		argIndex++
	}

	if request.Capacity > 0 {
		// Check if new capacity is less than current filled count
		var currentFilled int
		filledQuery := `SELECT filled FROM events_schema.events WHERE event_id = $1`
		err := er.db.db.QueryRow(filledQuery, eventID).Scan(&currentFilled)
		if err != nil {
			return nil, fmt.Errorf("failed to get current filled count: %v", err)
		}

		if request.Capacity < currentFilled {
			return nil, fmt.Errorf("cannot set capacity (%d) lower than current bookings (%d)", request.Capacity, currentFilled)
		}

		setParts = append(setParts, fmt.Sprintf("capacity = $%d", argIndex))
		args = append(args, request.Capacity)
		argIndex++
	}

	if len(setParts) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	// Add updated_at timestamp
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add WHERE clause parameters
	args = append(args, eventID)

	updateQuery := fmt.Sprintf("UPDATE events_schema.events SET %s WHERE event_id = $%d", strings.Join(setParts, ", "), argIndex)

	_, err = er.db.db.Exec(updateQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update event: %v", err)
	}

	// Return the updated event
	return er.GetEventByID(eventID)
}

// DeleteEvent deletes an event (organizer functionality)
func (er *EventRepo) DeleteEvent(eventID int, organizerID int) error {
	// First verify the organizer owns this event
	ownerQuery := `SELECT COUNT(*) FROM events_schema.events WHERE event_id = $1 AND organizer_id = $2`
	var count int
	err := er.db.db.QueryRow(ownerQuery, eventID, organizerID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to verify event ownership: %v", err)
	}

	if count == 0 {
		return fmt.Errorf("event not found or you don't have permission to delete it")
	}

	// Delete the event (CASCADE will handle user bookings)
	deleteQuery := `DELETE FROM events_schema.events WHERE event_id = $1`
	_, err = er.db.db.Exec(deleteQuery, eventID)
	if err != nil {
		return fmt.Errorf("failed to delete event: %v", err)
	}

	return nil
}
