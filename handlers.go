package grpcerrors

import (
	"github.com/srvc/fail"
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

// FailHandlerFunc is a function that called by interceptors when specified application erorrs are detected.
type FailHandlerFunc func(context.Context, *fail.Error) error

type failHandler struct {
	f FailHandlerFunc
}

func (h *failHandler) HandleUnaryServerError(c context.Context, req interface{}, info *grpc.UnaryServerInfo, err error) error {
	return h.handleError(c, err)
}

func (h *failHandler) HandleStreamServerError(c context.Context, req interface{}, resp interface{}, info *grpc.StreamServerInfo, err error) error {
	return h.handleError(c, err)
}

func (h *failHandler) handleError(c context.Context, err error) error {
	fErr := fail.Unwrap(err)
	if fErr != nil {
		return h.f(c, fErr)
	}
	return err
}

// WithFailHandler returns a new error handler function for handling errors wrapped with fail.Error.
func WithFailHandler(f FailHandlerFunc) interface {
	UnaryServerErrorHandler
	StreamServerErrorHandler
} {
	return &failHandler{f: f}
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
	fErr := fail.Unwrap(err)
	if fErr == nil {
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
func WithReportableErrorHandler(f FailHandlerFunc) interface {
	UnaryServerErrorHandler
	StreamServerErrorHandler
} {
	return WithFailHandler(func(c context.Context, err *fail.Error) error {
		if err.Ignorable {
			return err
		}
		return f(c, err)
	})
}

// CodeMap maps any status codes to gRPC's `codes.Code`s.
type CodeMap map[interface{}]codes.Code

// WithStatusCodeMap returns a new error handler function for mapping status codes to gRPC's one.
func WithStatusCodeMap(m CodeMap) interface {
	UnaryServerErrorHandler
	StreamServerErrorHandler
} {
	return WithFailHandler(func(c context.Context, err *fail.Error) error {
		if c, ok := m[err.Code]; ok {
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
	return WithFailHandler(func(c context.Context, err *fail.Error) error {
		return status.Error(mapFn(err.Code), err.Error())
	})
}

// WithGrpcStatusUnwrapper returns unwrapped error if this has a gRPC status.
func WithGrpcStatusUnwrapper() interface {
	UnaryServerErrorHandler
	StreamServerErrorHandler
} {
	return WithFailHandler(func(c context.Context, err *fail.Error) error {
		if _, ok := status.FromError(err.Err); ok {
			return err.Err
		}
		return err
	})
}
