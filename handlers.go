package grpcerrors

import (
	"github.com/creasty/apperrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorHandlerFunc is a function that called by interceptors when specified erorrs are detected.
type ErrorHandlerFunc func(error) error

// AppErrorHandlerFunc is a function that called by interceptors when specified application erorrs are detected.
type AppErrorHandlerFunc func(*apperrors.Error) error

// WithAppErrorHandler returns a new error handler function for handling errors wrapped with apperrors.
func WithAppErrorHandler(f AppErrorHandlerFunc) ErrorHandlerFunc {
	return func(err error) error {
		appErr := apperrors.Unwrap(err)
		if appErr != nil {
			return f(appErr)
		}
		return err
	}
}

// WithNotWrappedErrorHandler returns a new error handler function for handling not wrapped errors.
func WithNotWrappedErrorHandler(f ErrorHandlerFunc) ErrorHandlerFunc {
	return func(err error) error {
		appErr := apperrors.Unwrap(err)
		if appErr == nil {
			return f(err)
		}
		return err
	}
}

// WithReportableErrorHandler returns a new error handler function for handling errors annotated with the reportability.
func WithReportableErrorHandler(f AppErrorHandlerFunc) ErrorHandlerFunc {
	return WithAppErrorHandler(func(err *apperrors.Error) error {
		if err.Report {
			return f(err)
		}
		return err
	})
}

// WithStatusCodeMapper returns a new error handler function for mapping status codes to gRPC's one.
func WithStatusCodeMapper(m map[int]codes.Code) ErrorHandlerFunc {
	return WithAppErrorHandler(func(err *apperrors.Error) error {
		newCode := codes.Internal
		if c, ok := m[err.StatusCode]; ok {
			newCode = c
		}
		return status.Error(newCode, err.Error())
	})
}
