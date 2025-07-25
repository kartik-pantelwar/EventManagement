package middleware

import (
	errorhandling "authservice/src/pkg/error_handling"
	"authservice/src/pkg/utilities"
	"context"
	"net/http"
)

// Authenticate middleware now supports role-based authentication (customer/organizer)
func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("at")
		if err != nil {
			errorhandling.HandleError(w, "Missing Authorization Token", http.StatusUnauthorized)
			return
		}

		claims, err := utilities.ValidateJWT(cookie.Value)
		if err != nil {
			errorhandling.HandleError(w, "Invalid Authorization Token", http.StatusUnauthorized)
			return
		}

		// Expect claims to have Uid and Role fields
		ctx := context.WithValue(r.Context(), "user", claims.Uid)
		ctx = context.WithValue(ctx, "role", claims.Role)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
