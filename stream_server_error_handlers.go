package grpcerrors

import (
	"github.com/creasty/apperrors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// StreamServerAppErrorHandlerFunc is a function that called by stream server interceptors when specified application erorrs are detected.
type StreamServerAppErrorHandlerFunc func(context.Context, interface{}, interface{}, *grpc.StreamServerInfo, *apperrors.Error) error

type streamServerAppErrorHandler struct {
	f StreamServerAppErrorHandlerFunc
}

func (h *streamServerAppErrorHandler) HandleStreamServerError(c context.Context, req interface{}, resp interface{}, info *grpc.StreamServerInfo, err error) error {
	appErr := apperrors.Unwrap(err)
	if appErr != nil {
		return h.f(c, req, resp, info, appErr)
	}
	return err
}

// WithStreamServerAppErrorHandler returns a new error handler for stream servers for handling errors wrapped with apperrors.
func WithStreamServerAppErrorHandler(f StreamServerAppErrorHandlerFunc) StreamServerErrorHandler {
	return &streamServerAppErrorHandler{f: f}
}

// WithStreamServerReportableErrorHandler returns a new error handler for stream servers for handling errors annotated with the reportability.
func WithStreamServerReportableErrorHandler(f StreamServerAppErrorHandlerFunc) StreamServerErrorHandler {
	return WithStreamServerAppErrorHandler(func(c context.Context, req interface{}, resp interface{}, info *grpc.StreamServerInfo, err *apperrors.Error) error {
		if err.Report {
			return f(c, req, resp, info, err)
		}
		return err
	})
}
