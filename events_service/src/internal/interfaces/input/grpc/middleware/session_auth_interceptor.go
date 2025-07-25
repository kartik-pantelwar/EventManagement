package middleware

import (
	"context"
	"eventservice/src/pkg/response"
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

// SessionIDMiddleware extracts session_id from the request header and validates it via gRPC
func (m *SessionAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.Header.Get("Session-Id")
		if sessionID == "" {
			response.WriteError(w, http.StatusUnauthorized, "Missing session ID")
			return
		}

		// gRPC call to auth service
		resp, err := m.GrpcClient.ValidateSession(context.Background(), &pb.ValidateSessionRequest{
			SessionId: sessionID,
		})
		if err != nil || !resp.Valid {
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

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
