package grpcerrors

import (
	"github.com/srvc/fail/v4"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// UnaryServerFailHandlerFunc is a function that called by unary server interceptors when specified application erorrs are detected.
type UnaryServerFailHandlerFunc func(context.Context, interface{}, *grpc.UnaryServerInfo, *fail.Error) error

type unaryServerFailHandler struct {
	f UnaryServerFailHandlerFunc
}

func (h *unaryServerFailHandler) HandleUnaryServerError(c context.Context, req interface{}, info *grpc.UnaryServerInfo, err error) error {
	fErr := fail.Unwrap(err)
	if fErr != nil {
		return h.f(c, req, info, fErr)
	}
	return err
}

// WithUnaryServerFailHandler returns a new error handler for unary servers for handling errors wrapped with fail.Error.
func WithUnaryServerFailHandler(f UnaryServerFailHandlerFunc) UnaryServerErrorHandler {
	return &unaryServerFailHandler{f: f}
}

// WithUnaryServerReportableErrorHandler returns a new error handler for unary servers for handling errors annotated with the reportability.
func WithUnaryServerReportableErrorHandler(f UnaryServerFailHandlerFunc) UnaryServerErrorHandler {
	return WithUnaryServerFailHandler(func(c context.Context, req interface{}, info *grpc.UnaryServerInfo, err *fail.Error) error {
		if err.Ignorable {
			return err
		}
		return f(c, req, info, err)
	})
}
