package routes

import (
	customerhandler "authservice/src/internal/interfaces/input/rest/handler/customer"
	organizerhandler "authservice/src/internal/interfaces/input/rest/handler/organizer"
	"authservice/src/internal/interfaces/input/rest/middleware"

	"net/http"

	"github.com/go-chi/chi/v5"
)

func InitRoutes(
	customerHandler *customerhandler.CustomerHandler,
	organizerHandler *organizerhandler.OrganizerHandler) http.Handler {
	router := chi.NewRouter()

	// Customer routes
	router.Route("/customers", func(r chi.Router) {
		r.Post("/register", customerHandler.Register)
		r.Post("/login", customerHandler.Login)
		r.Post("/refresh", customerHandler.Refresh)
		r.Post("/verify-otp", customerHandler.VerifyOTP) // OTP verification route

		// Protected customer routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate)
			r.Get("/profile", customerHandler.Profile)
			r.Post("/logout", customerHandler.LogOut)
		})
	})

	// Organizer routes
	router.Route("/organizers", func(r chi.Router) {
		r.Post("/register", organizerHandler.Register)
		r.Post("/login", organizerHandler.Login)
		r.Post("/refresh", organizerHandler.Refresh)
		r.Post("/verify-otp", organizerHandler.VerifyOTP) // OTP verification route

		// Protected organizer routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate)
			r.Get("/profile", organizerHandler.Profile)
			r.Post("/logout", organizerHandler.LogOut)
		})
	})

	return router
}
