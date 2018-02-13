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

type appErrorHandler struct {
	f AppErrorHandlerFunc
}

func (h *appErrorHandler) HandleUnaryServerError(c context.Context, req interface{}, info *grpc.UnaryServerInfo, err error) error {
	return h.handleError(c, err)
}

func (h *appErrorHandler) HandleStreamServerError(c context.Context, req interface{}, resp interface{}, info *grpc.StreamServerInfo, err error) error {
	return h.handleError(c, err)
}

func (h *appErrorHandler) handleError(c context.Context, err error) error {
	appErr := apperrors.Unwrap(err)
	if appErr != nil {
		return h.f(c, appErr)
	}
	return err
}

// WithAppErrorHandler returns a new error handler function for handling errors wrapped with apperrors.
func WithAppErrorHandler(f AppErrorHandlerFunc) interface {
	UnaryServerErrorHandler
	StreamServerErrorHandler
} {
	return &appErrorHandler{f: f}
}

type notWrappedHandler struct {
	f ErrorHandlerFunc
}

func (h *notWrappedHandler) HandleUnaryServerError(c context.Context, req interface{}, info *grpc.UnaryServerInfo, err error) error {
	return h.handleError(c, err)
}

func (h *notWrappedHandler) HandleStreamServerError(c context.Context, req interface{}, resp interface{}, info *grpc.StreamServerInfo, err error) error {
	return h.handleError(c, err)
}

func (h *notWrappedHandler) handleError(c context.Context, err error) error {
	appErr := apperrors.Unwrap(err)
	if appErr == nil {
		return h.f(c, err)
	}
	return err
}

// WithNotWrappedErrorHandler returns a new error handler function for handling not wrapped errors.
func WithNotWrappedErrorHandler(f ErrorHandlerFunc) interface {
	UnaryServerErrorHandler
	StreamServerErrorHandler
} {
	return &notWrappedHandler{f: f}
}

// WithReportableErrorHandler returns a new error handler function for handling errors annotated with the reportability.
func WithReportableErrorHandler(f AppErrorHandlerFunc) interface {
	UnaryServerErrorHandler
	StreamServerErrorHandler
} {
	return WithAppErrorHandler(func(c context.Context, err *apperrors.Error) error {
		if err.Report {
			return f(c, err)
		}
		return err
	})
}

// WithStatusCodeMap returns a new error handler function for mapping status codes to gRPC's one.
func WithStatusCodeMap(m map[int]codes.Code) interface {
	UnaryServerErrorHandler
	StreamServerErrorHandler
} {
	return WithAppErrorHandler(func(c context.Context, err *apperrors.Error) error {
		newCode := codes.Unknown
		if s, ok := status.FromError(err.Err); ok {
			newCode = s.Code()
		} else if c, ok := m[err.StatusCode]; ok {
			newCode = c
		}
		return status.Error(newCode, err.Error())
	})
}

// WithStatusCodeMapper returns a new error handler function for mapping status codes to gRPC's one with given function.
func WithStatusCodeMapper(mapFn func(code int) codes.Code) interface {
	UnaryServerErrorHandler
	StreamServerErrorHandler
} {
	return WithAppErrorHandler(func(c context.Context, err *apperrors.Error) error {
		return status.Error(mapFn(err.StatusCode), err.Error())
	})
}
