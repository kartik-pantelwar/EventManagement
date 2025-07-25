package routes

import (
	"eventservice/src/internal/interfaces/input/grpc/middleware"
	"eventservice/src/internal/interfaces/input/rest/handler/event"
	"net/http"

	pb "eventservice/src/internal/interfaces/input/grpc/generated"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func InitRoutes(eventHandler *event.EventHandler, grpcClient pb.ValidationServiceClient) http.Handler {
	router := chi.NewRouter()

	// Add basic middleware
	router.Use(chimiddleware.Logger)
	router.Use(chimiddleware.Recoverer)
	router.Use(chimiddleware.RequestID)

	// CORS middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Session-Id")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Initialize session auth middleware
	sessionAuth := middleware.NewSessionAuthMiddleware(grpcClient)

	// Public routes (no authentication required)
	router.Route("/api/v1", func(r chi.Router) {
		// Health check
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Events Service is healthy"))
		})

		// Public events routes (read-only)
		r.Route("/events", func(r chi.Router) {
			r.Get("/", eventHandler.GetAllEvents) // Get all events with filters
			r.Get("/{id}", eventHandler.GetEvent) // Get specific event
		})
	})

	// Protected routes (authentication required)
	router.Route("/api/v1", func(r chi.Router) {
		r.Use(sessionAuth.Middleware) // Apply session validation

		// General authenticated routes (both organizers and customers)
		r.Route("/events", func(r chi.Router) {
			// Customer routes (joining/leaving events)
			r.With(sessionAuth.CustomerOnly).Post("/{id}/join", eventHandler.JoinEvent)
			r.With(sessionAuth.CustomerOnly).Delete("/{id}/leave", eventHandler.LeaveEvent)
		})

		// Customer-specific routes
		r.Route("/user", func(r chi.Router) {
			r.Use(sessionAuth.CustomerOnly)
			r.Get("/bookings", eventHandler.GetMyBookings) // Get user's booked events
		})

		// Organizer-specific routes
		r.Route("/organizer", func(r chi.Router) {
			r.Use(sessionAuth.OrganizerOnly)

			// Event management
			r.Post("/events", eventHandler.CreateEvent)        // Create event
			r.Get("/events", eventHandler.GetMyEvents)         // Get organizer's events
			r.Put("/events/{id}", eventHandler.UpdateEvent)    // Update event
			r.Delete("/events/{id}", eventHandler.DeleteEvent) // Delete event

			// Event participants
			r.Get("/events/{id}/participants", eventHandler.GetEventParticipants) // Get event participants
		})
	})

	return router
}
