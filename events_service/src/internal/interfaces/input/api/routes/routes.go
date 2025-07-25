package routes

import (
	"eventservice/src/internal/interfaces/input/grpc/middleware"
	"eventservice/src/internal/interfaces/input/rest/handler/event"
	"net/http"

	pb "eventservice/src/internal/interfaces/input/grpc/generated"

	"github.com/go-chi/chi/v5"
)

func InitRoutes(eventHandler *event.EventHandler, grpcClient pb.ValidationServiceClient) http.Handler {
	router := chi.NewRouter()

	// Initialize session auth middleware
	sessionAuth := middleware.NewSessionAuthMiddleware(grpcClient)

	// API routes
	router.Route("/api/v1", func(r chi.Router) {
		// Public events routes
		r.Route("/events", func(r chi.Router) {
			r.Get("/", eventHandler.GetAllEvents) // Get all events with filters
			r.Get("/{id}", eventHandler.GetEvent) // Get specific event

			// Protected event routes (authentication required)
			r.Group(func(r chi.Router) {
				r.Use(sessionAuth.Middleware) // Apply session validation
				// Customer routes 
				r.With(sessionAuth.CustomerOnly).Post("/{id}/join", eventHandler.JoinEvent)
				r.With(sessionAuth.CustomerOnly).Delete("/{id}/leave", eventHandler.LeaveEvent)
			})
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(sessionAuth.Middleware)

			// Customerroutes
			r.Route("/user", func(r chi.Router) {
				r.Use(sessionAuth.CustomerOnly)
				r.Get("/bookings", eventHandler.GetMyBookings) // Get user's booked events
			})

			// Organizer-specific routes
			r.Route("/organizer", func(r chi.Router) {
				r.Use(sessionAuth.OrganizerOnly)
				r.Post("/events", eventHandler.CreateEvent)        // Create event
				r.Get("/events", eventHandler.GetMyEvents)         // Get organizer's events
				r.Put("/events/{id}", eventHandler.UpdateEvent)    // Update event
				r.Delete("/events/{id}", eventHandler.DeleteEvent) // Delete event

				// Event participants
				r.Get("/events/{id}/participants", eventHandler.GetEventParticipants) // Get event participants
			})
		})
	})

	return router
}
