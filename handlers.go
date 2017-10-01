package grpcerrors

import (
	"github.com/creasty/apperrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorHandlerFunc func(error) error
type AppErrorHandlerFunc func(*apperrors.Error) error

func HandleNotWrappedError(f ErrorHandlerFunc) ErrorHandlerFunc {
	return func(err error) error {
		appErr := apperrors.Unwrap(err)
		if appErr == nil {
			return f(err)
		}
		return err
	}
}

func Report(f AppErrorHandlerFunc) ErrorHandlerFunc {
	return handleAppError(func(err *apperrors.Error) error {
		if err.Report {
			return f(err)
		}
		return err
	})
}

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
