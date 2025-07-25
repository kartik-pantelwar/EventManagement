package client

import (
	pb "eventservice/src/internal/interfaces/input/grpc/generated"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewSessionValidatorClient(addr string) (pb.ValidationServiceClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return pb.NewValidationServiceClient(conn), nil
}
