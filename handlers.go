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

// HandleNotWrappedError returns a new error handler function for handling not wrapped errors.
func HandleNotWrappedError(f ErrorHandlerFunc) ErrorHandlerFunc {
	return func(err error) error {
		appErr := apperrors.Unwrap(err)
		if appErr == nil {
			return f(err)
		}
		return err
	}
}

// Report returns a new error handler function for handling errors annotated with the reportability.
func Report(f AppErrorHandlerFunc) ErrorHandlerFunc {
	return handleAppError(func(err *apperrors.Error) error {
		if err.Report {
			return f(err)
		}
		return err
	})
}

// MapStatusCode returns a new error handler function for mapping status codes to gRPC's one.
func MapStatusCode(m map[int]codes.Code) ErrorHandlerFunc {
	return handleAppError(func(err *apperrors.Error) error {
		newCode := codes.Internal
		if c, ok := m[err.StatusCode]; ok {
			newCode = c
		}
		return status.Error(newCode, err.Error())
	})
}

func handleAppError(f AppErrorHandlerFunc) ErrorHandlerFunc {
	return func(err error) error {
		appErr := apperrors.Unwrap(err)
		if appErr != nil {
			return f(appErr)
		}
		return err
	}
}
