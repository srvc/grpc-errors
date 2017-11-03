package grpcerrors

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// UnaryServerInterceptor returns a new unary server interceptor to handle errors
func UnaryServerInterceptor(handlers ...UnaryServerErrorHandler) grpc.UnaryServerInterceptor {
	errHandler := composeUnaryServerErrorHandlers(handlers)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		return resp, errHandler.HandleUnaryServerError(ctx, req, info, err)
	}
}

// StreamServerInterceptor returns a new streaming server interceptor to handle errors
func StreamServerInterceptor(handlers ...StreamServerErrorHandler) grpc.StreamServerInterceptor {
	errHandler := composeStreamServerErrorHandlers(handlers)
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		newStream := &recordableServerStream{ServerStream: stream}
		err := handler(srv, newStream)
		return errHandler.HandleStreamServerError(
			newStream.Context(),
			newStream.request,
			newStream.response,
			info,
			err,
		)
	}
}

type recordableServerStream struct {
	grpc.ServerStream
	request  interface{}
	response interface{}
}

func (s *recordableServerStream) SendMsg(m interface{}) error {
	s.response = m
	return s.ServerStream.SendMsg(m)
}

func (s *recordableServerStream) RecvMsg(m interface{}) error {
	s.request = m
	return s.ServerStream.RecvMsg(m)
}
