package middleware

import (
	"context"
	"eventservice/src/pkg/response"
	"log"
	"net/http"
	"strconv"

	pb "eventservice/src/internal/interfaces/input/grpc/generated"
)

type SessionAuthMiddleware struct {
	GrpcClient pb.ValidationServiceClient
}

func NewSessionAuthMiddleware(grpcClient pb.ValidationServiceClient) *SessionAuthMiddleware {
	return &SessionAuthMiddleware{
		GrpcClient: grpcClient,
	}
}

// SessionIDMiddleware extracts session_id from the request cookie and validates it via gRPC
func (m *SessionAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("DEBUG: Processing request to %s", r.URL.Path)

		// Try to get session ID from cookie first (preferred method)
		sessionID := ""
		if cookie, err := r.Cookie("sess"); err == nil {
			sessionID = cookie.Value
			log.Printf("DEBUG: Found session ID in cookie: %s", sessionID)
		} else {
			log.Printf("DEBUG: No 'sess' cookie found: %v", err)
		}

		// Fallback to header if cookie not found (for backward compatibility)
		if sessionID == "" {
			sessionID = r.Header.Get("Session-Id")
			if sessionID != "" {
				log.Printf("DEBUG: Found session ID in header: %s", sessionID)
			} else {
				log.Printf("DEBUG: No 'Session-Id' header found")
			}
		}

		if sessionID == "" {
			log.Printf("DEBUG: No session ID found in cookies or headers")
			response.WriteError(w, http.StatusUnauthorized, "Missing session ID")
			return
		}

		log.Printf("DEBUG: Making gRPC call to validate session: %s", sessionID)

		// gRPC call to auth service
		resp, err := m.GrpcClient.ValidateSession(context.Background(), &pb.ValidateSessionRequest{
			SessionId: sessionID,
		})
		if err != nil {
			log.Printf("DEBUG: gRPC call failed: %v", err)
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		if !resp.Valid {
			log.Printf("DEBUG: Session validation failed: %s", resp.Error)
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		log.Printf("DEBUG: Session validated successfully for user: %s, role: %s", resp.UserId, resp.Role)

		// Convert user_id from string to int
		userID, err := strconv.Atoi(resp.UserId)
		if err != nil {
			response.WriteError(w, http.StatusUnauthorized, "Invalid user ID")
			return
		}

		// Add user_id and role to context
		ctx := context.WithValue(r.Context(), "userID", userID)
		ctx = context.WithValue(ctx, "role", resp.Role)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// OrganizerOnly middleware ensures only organizers can access certain endpoints
func (m *SessionAuthMiddleware) OrganizerOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value("role").(string)
		if !ok || role != "organizer" {
			response.WriteError(w, http.StatusForbidden, "Organizer access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// CustomerOnly middleware ensures only customers can access certain endpoints
func (m *SessionAuthMiddleware) CustomerOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value("role").(string)
		if !ok || role != "customer" {
			response.WriteError(w, http.StatusForbidden, "Customer access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
