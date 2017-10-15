package grpcerrors

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// UnaryServerInterceptor returns a new unary server interceptor to handle errors
func UnaryServerInterceptor(funcs ...ErrorHandlerFunc) grpc.UnaryServerInterceptor {
	handleError := composeHandlers(funcs)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		return resp, handleError(ctx, err)
	}
}

// StreamServerInterceptor returns a new streaming server interceptor to handle errors
func StreamServerInterceptor(funcs ...ErrorHandlerFunc) grpc.StreamServerInterceptor {
	handleError := composeHandlers(funcs)
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handleError(stream.Context(), handler(srv, stream))
	}
}
