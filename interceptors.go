package grpcerrors

import (
	"context"

	"google.golang.org/grpc"
)

// UnaryServerInterceptor returns a new unary server interceptor to handle errors
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new streaming server interceptor to handle errors
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, stream)
	}
}
