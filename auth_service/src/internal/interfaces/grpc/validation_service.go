package grpc

import (
	"authservice/src/internal/adaptors/persistance"
	"authservice/src/internal/interfaces/grpc/generated"
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
)

type ValidationServer struct {
	generated.UnimplementedValidationServiceServer
	sessionRepo persistance.SessionRepo
}

func NewValidationServer(sessionRepo persistance.SessionRepo) *ValidationServer {
	return &ValidationServer{
		sessionRepo: sessionRepo,
	}
}

// ValidateSession implements the gRPC validation service
func (vs *ValidationServer) ValidateSession(ctx context.Context, req *generated.ValidateSessionRequest) (*generated.ValidateSessionResponse, error) {
	log.Printf("Validating session: %s", req.SessionId)

	// Get the session using the existing GetSession method
	session, err := vs.sessionRepo.GetSession(req.SessionId)
	if err != nil {
		log.Printf("Session validation failed: %v", err)
		return &generated.ValidateSessionResponse{
			Valid: false,
			Error: "Invalid session",
		}, nil
	}

	// Check if session is expired
	if session.ExpiresAt.Before(time.Now()) {
		log.Printf("Session expired for user: %d", session.Uid)
		return &generated.ValidateSessionResponse{
			Valid: false,
			Error: "Session expired",
		}, nil
	}

	// Get user role from database
	role, err := vs.sessionRepo.GetUserRole(session.Uid)
	if err != nil {
		log.Printf("Failed to get user role: %v", err)
		return &generated.ValidateSessionResponse{
			Valid: false,
			Error: "Failed to get user role",
		}, nil
	}

	log.Printf("Session validated successfully for user: %d, role: %s", session.Uid, role)

	return &generated.ValidateSessionResponse{
		Valid:  true,
		UserId: fmt.Sprintf("%d", session.Uid),
		Role:   role,
		Error:  "",
	}, nil
}

// StartGRPCServer starts the gRPC server for validation service
func StartGRPCServer(port string, sessionRepo persistance.SessionRepo) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	validationServer := NewValidationServer(sessionRepo)

	generated.RegisterValidationServiceServer(grpcServer, validationServer)

	log.Printf("gRPC server starting on port %s", port)
	return grpcServer.Serve(listener)
}
