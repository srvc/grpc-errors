package grpcerrors

import (
	"github.com/izumin5210/fail"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// UnaryServerFailHandlerFunc is a function that called by unary server interceptors when specified application erorrs are detected.
type UnaryServerFailHandlerFunc func(context.Context, interface{}, *grpc.UnaryServerInfo, *fail.Error) error

type unaryServerFailHandler struct {
	f UnaryServerFailHandlerFunc
}

func (h *unaryServerFailHandler) HandleUnaryServerError(c context.Context, req interface{}, info *grpc.UnaryServerInfo, err error) error {
	appErr := fail.Unwrap(err)
	if appErr != nil {
		return h.f(c, req, info, appErr)
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
		if err.Report {
			return f(c, req, info, err)
		}
		return err
	})
}
