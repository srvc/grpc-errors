package grpcerrors

import (
	"github.com/izumin5210/fail"
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
type AppErrorHandlerFunc func(context.Context, *fail.Error) error

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
	appErr := fail.Unwrap(err)
	if appErr != nil {
		return h.f(c, appErr)
	}
	return err
}

// WithAppErrorHandler returns a new error handler function for handling errors wrapped with fail.Error.
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
	appErr := fail.Unwrap(err)
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
	return WithAppErrorHandler(func(c context.Context, err *fail.Error) error {
		if err.Report {
			return f(c, err)
		}
		return err
	})
}

// CodeMap maps any status codes to gRPC's `codes.Code`s.
type CodeMap map[interface{}]codes.Code

// WithStatusCodeMap returns a new error handler function for mapping status codes to gRPC's one.
func WithStatusCodeMap(m CodeMap) interface {
	UnaryServerErrorHandler
	StreamServerErrorHandler
} {
	return WithAppErrorHandler(func(c context.Context, err *fail.Error) error {
		if c, ok := m[err.StatusCode]; ok {
			return status.Error(c, err.Error())
		}
		return err
	})
}

// CodeMapFunc returns gRPC's `codes.Code`s from given any codes.
type CodeMapFunc func(code interface{}) codes.Code

// WithStatusCodeMapper returns a new error handler function for mapping status codes to gRPC's one with given function.
func WithStatusCodeMapper(mapFn CodeMapFunc) interface {
	UnaryServerErrorHandler
	StreamServerErrorHandler
} {
	return WithAppErrorHandler(func(c context.Context, err *fail.Error) error {
		return status.Error(mapFn(err.StatusCode), err.Error())
	})
}

// WithGrpcStatusUnwrapper returns unwrapped error if this has a gRPC status.
func WithGrpcStatusUnwrapper() interface {
	UnaryServerErrorHandler
	StreamServerErrorHandler
} {
	return WithAppErrorHandler(func(c context.Context, err *fail.Error) error {
		if _, ok := status.FromError(err.Err); ok {
			return err.Err
		}
		return err
	})
}
