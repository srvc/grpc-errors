package grpcerrors

import (
	"github.com/creasty/apperrors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerErrorHandler is the interface that can handle errors on a gRPC unary server
type UnaryServerErrorHandler interface {
	HandleUnaryServerError(context.Context, interface{}, *grpc.UnaryServerInfo, error) error
}

// StreamServerErrorHandler is the interface that can handle errors on a gRPC stream server
type StreamServerErrorHandler interface {
	HandleStreamServerError(context.Context, interface{}, interface{}, *grpc.StreamServerInfo, error) error
}

// ErrorHandlerFunc is a function that called by interceptors when specified erorrs are detected.
type ErrorHandlerFunc func(context.Context, error) error

// AppErrorHandlerFunc is a function that called by interceptors when specified application erorrs are detected.
type AppErrorHandlerFunc func(context.Context, *apperrors.Error) error

// WithAppErrorHandler returns a new error handler function for handling errors wrapped with apperrors.
func WithAppErrorHandler(f AppErrorHandlerFunc) ErrorHandlerFunc {
	return func(c context.Context, err error) error {
		appErr := apperrors.Unwrap(err)
		if appErr != nil {
			return f(c, appErr)
		}
		return err
	}
}

// WithNotWrappedErrorHandler returns a new error handler function for handling not wrapped errors.
func WithNotWrappedErrorHandler(f ErrorHandlerFunc) ErrorHandlerFunc {
	return func(c context.Context, err error) error {
		appErr := apperrors.Unwrap(err)
		if appErr == nil {
			return f(c, err)
		}
		return err
	}
}

// WithReportableErrorHandler returns a new error handler function for handling errors annotated with the reportability.
func WithReportableErrorHandler(f AppErrorHandlerFunc) ErrorHandlerFunc {
	return WithAppErrorHandler(func(c context.Context, err *apperrors.Error) error {
		if err.Report {
			return f(c, err)
		}
		return err
	})
}

// WithStatusCodeMap returns a new error handler function for mapping status codes to gRPC's one.
func WithStatusCodeMap(m map[int]codes.Code) ErrorHandlerFunc {
	return WithAppErrorHandler(func(c context.Context, err *apperrors.Error) error {
		newCode := codes.Internal
		if c, ok := m[err.StatusCode]; ok {
			newCode = c
		}
		return status.Error(newCode, err.Error())
	})
}

// WithStatusCodeMapper returns a new error handler function for mapping status codes to gRPC's one with given function.
func WithStatusCodeMapper(mapFn func(code int) codes.Code) ErrorHandlerFunc {
	return WithAppErrorHandler(func(c context.Context, err *apperrors.Error) error {
		return status.Error(mapFn(err.StatusCode), err.Error())
	})
}
