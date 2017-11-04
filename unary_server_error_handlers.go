package grpcerrors

import (
	"github.com/creasty/apperrors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// UnaryServerAppErrorHandlerFunc is a function that called by unary server interceptors when specified application erorrs are detected.
type UnaryServerAppErrorHandlerFunc func(context.Context, interface{}, *grpc.UnaryServerInfo, *apperrors.Error) error

type unaryServerAppErrorHandler struct {
	f UnaryServerAppErrorHandlerFunc
}

func (h *unaryServerAppErrorHandler) HandleUnaryServerError(c context.Context, req interface{}, info *grpc.UnaryServerInfo, err error) error {
	appErr := apperrors.Unwrap(err)
	if appErr != nil {
		return h.f(c, req, info, appErr)
	}
	return err
}

// WithUnaryServerAppErrorHandler returns a new error handler for unary servers for handling errors wrapped with apperrors.
func WithUnaryServerAppErrorHandler(f UnaryServerAppErrorHandlerFunc) UnaryServerErrorHandler {
	return &unaryServerAppErrorHandler{f: f}
}

// WithUnaryServerReportableErrorHandler returns a new error handler for unary servers for handling errors annotated with the reportability.
func WithUnaryServerReportableErrorHandler(f UnaryServerAppErrorHandlerFunc) UnaryServerErrorHandler {
	return WithUnaryServerAppErrorHandler(func(c context.Context, req interface{}, info *grpc.UnaryServerInfo, err *apperrors.Error) error {
		if err.Report {
			return f(c, req, info, err)
		}
		return err
	})
}
