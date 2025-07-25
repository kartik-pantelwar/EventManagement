// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.12.4
// source: auth.proto

package generated

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	ValidationService_ValidateSession_FullMethodName = "/session.ValidationService/ValidateSession"
)

// ValidationServiceClient is the client API for ValidationService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ValidationServiceClient interface {
	ValidateSession(ctx context.Context, in *ValidateSessionRequest, opts ...grpc.CallOption) (*ValidateSessionResponse, error)
}

type validationServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewValidationServiceClient(cc grpc.ClientConnInterface) ValidationServiceClient {
	return &validationServiceClient{cc}
}

func (c *validationServiceClient) ValidateSession(ctx context.Context, in *ValidateSessionRequest, opts ...grpc.CallOption) (*ValidateSessionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ValidateSessionResponse)
	err := c.cc.Invoke(ctx, ValidationService_ValidateSession_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ValidationServiceServer is the server API for ValidationService service.
// All implementations must embed UnimplementedValidationServiceServer
// for forward compatibility.
type ValidationServiceServer interface {
	ValidateSession(context.Context, *ValidateSessionRequest) (*ValidateSessionResponse, error)
	mustEmbedUnimplementedValidationServiceServer()
}

// UnimplementedValidationServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedValidationServiceServer struct{}

func (UnimplementedValidationServiceServer) ValidateSession(context.Context, *ValidateSessionRequest) (*ValidateSessionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateSession not implemented")
}
func (UnimplementedValidationServiceServer) mustEmbedUnimplementedValidationServiceServer() {}
func (UnimplementedValidationServiceServer) testEmbeddedByValue()                           {}

// UnsafeValidationServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ValidationServiceServer will
// result in compilation errors.
type UnsafeValidationServiceServer interface {
	mustEmbedUnimplementedValidationServiceServer()
}

func RegisterValidationServiceServer(s grpc.ServiceRegistrar, srv ValidationServiceServer) {
	// If the following call pancis, it indicates UnimplementedValidationServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ValidationService_ServiceDesc, srv)
}

func _ValidationService_ValidateSession_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidateSessionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ValidationServiceServer).ValidateSession(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ValidationService_ValidateSession_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ValidationServiceServer).ValidateSession(ctx, req.(*ValidateSessionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ValidationService_ServiceDesc is the grpc.ServiceDesc for ValidationService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ValidationService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "session.ValidationService",
	HandlerType: (*ValidationServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ValidateSession",
			Handler:    _ValidationService_ValidateSession_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "auth.proto",
}
