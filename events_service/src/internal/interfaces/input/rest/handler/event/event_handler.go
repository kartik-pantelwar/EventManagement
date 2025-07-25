package event

import (
	"encoding/json"
	"eventservice/src/internal/core"
	eventservice "eventservice/src/internal/usecase/event"
	"eventservice/src/pkg/response"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type EventHandler struct {
	eventService eventservice.Service
}

func NewEventHandler(es eventservice.Service) *EventHandler {
	return &EventHandler{eventService: es}
}

// CreateEvent handles POST /events
func (eh *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	// Get organizer ID from context (set by auth middleware)
	organizerID, ok := r.Context().Value("userID").(int)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var request core.CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Basic validation
	if request.EventName == "" || request.Place == "" ||
		request.EventDate == "" || request.StartTime == "" || request.EndTime == "" || request.Capacity <= 0 {
		response.WriteError(w, http.StatusBadRequest, "All fields are required and capacity must be positive")
		return
	}

	eventResponse, err := eh.eventService.CreateEvent(&request, organizerID)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteSuccess(w, http.StatusCreated, "Event created successfully", eventResponse)
}

// GetEvent handles GET /events/{id}
func (eh *EventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	eventIDStr := chi.URLParam(r, "id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}

	event, err := eh.eventService.GetEventByID(eventID)
	if err != nil {
		response.WriteError(w, http.StatusNotFound, err.Error())
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Event retrieved successfully", event)
}

// GetAllEvents handles GET /events
func (eh *EventHandler) GetAllEvents(w http.ResponseWriter, r *http.Request) {
	// Build filters from query parameters
	filters := &core.EventFilters{
		Date:        r.URL.Query().Get("date"),
		Place:       r.URL.Query().Get("place"),
		OrganizerID: 0, // Will be set below if provided
	}

	// Parse organizer ID if provided
	if organizerIDStr := r.URL.Query().Get("organizer_id"); organizerIDStr != "" {
		if organizerID, err := strconv.Atoi(organizerIDStr); err == nil {
			filters.OrganizerID = organizerID
		}
	}

	// Parse organizer_id if provided
	if organizerIDStr := r.URL.Query().Get("organizer_id"); organizerIDStr != "" {
		if organizerID, err := strconv.Atoi(organizerIDStr); err == nil {
			filters.OrganizerID = organizerID
		}
	}

	events, err := eh.eventService.GetAllEvents(filters)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Events retrieved successfully", events)
}

// GetMyEvents handles GET /organizer/events (for organizers to see their events)
func (eh *EventHandler) GetMyEvents(w http.ResponseWriter, r *http.Request) {
	// Get organizer ID from context (set by auth middleware)
	organizerID, ok := r.Context().Value("userID").(int)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	events, err := eh.eventService.GetEventsByOrganizer(organizerID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Events retrieved successfully", events)
}

// UpdateEvent handles PUT /events/{id}
func (eh *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	// Get organizer ID from context (set by auth middleware)
	organizerID, ok := r.Context().Value("userID").(int)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	eventIDStr := chi.URLParam(r, "id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}

	var request core.UpdateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	eventResponse, err := eh.eventService.UpdateEvent(eventID, &request, organizerID)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Event updated successfully", eventResponse)
}

// DeleteEvent handles DELETE /events/{id}
func (eh *EventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	// Get organizer ID from context (set by auth middleware)
	organizerID, ok := r.Context().Value("userID").(int)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	eventIDStr := chi.URLParam(r, "id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}

	err = eh.eventService.DeleteEvent(eventID, organizerID)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Event deleted successfully", nil)
}

// JoinEvent handles POST /events/{id}/join
func (eh *EventHandler) JoinEvent(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	eventIDStr := chi.URLParam(r, "id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}

	// For now, use placeholder values for email and username
	// In production, these should come from the auth service via gRPC
	request := &core.JoinEventRequest{EventID: eventID}
	bookingResponse, err := eh.eventService.JoinEventWithRequest(userID, request)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Successfully joined event", bookingResponse)
}

// LeaveEvent handles DELETE /events/{id}/leave
func (eh *EventHandler) LeaveEvent(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	eventIDStr := chi.URLParam(r, "id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}

	err = eh.eventService.LeaveEvent(userID, eventID)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Successfully left event", nil)
}

// GetMyBookings handles GET /user/bookings (for users to see their joined events)
func (eh *EventHandler) GetMyBookings(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	events, err := eh.eventService.GetUserBookings(userID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Bookings retrieved successfully", events)
}

// GetEventParticipants handles GET /events/{id}/participants (for organizers)
func (eh *EventHandler) GetEventParticipants(w http.ResponseWriter, r *http.Request) {
	// Get organizer ID from context (set by auth middleware)
	organizerID, ok := r.Context().Value("userID").(int)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	eventIDStr := chi.URLParam(r, "id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}

	participants, err := eh.eventService.GetEventParticipants(eventID, organizerID)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Participants retrieved successfully", participants)
}
