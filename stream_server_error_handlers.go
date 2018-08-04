package grpcerrors

import (
	"github.com/izumin5210/fail"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// StreamServerFailHandlerFunc is a function that called by stream server interceptors when specified application erorrs are detected.
type StreamServerFailHandlerFunc func(context.Context, interface{}, interface{}, *grpc.StreamServerInfo, *fail.Error) error

type streamServerFailHandler struct {
	f StreamServerFailHandlerFunc
}

func (h *streamServerFailHandler) HandleStreamServerError(c context.Context, req interface{}, resp interface{}, info *grpc.StreamServerInfo, err error) error {
	fErr := fail.Unwrap(err)
	if fErr != nil {
		return h.f(c, req, resp, info, fErr)
	}
	return err
}

// WithStreamServerFailHandler returns a new error handler for stream servers for handling errors wrapped with fail.Error.
func WithStreamServerFailHandler(f StreamServerFailHandlerFunc) StreamServerErrorHandler {
	return &streamServerFailHandler{f: f}
}

// WithStreamServerReportableErrorHandler returns a new error handler for stream servers for handling errors annotated with the reportability.
func WithStreamServerReportableErrorHandler(f StreamServerFailHandlerFunc) StreamServerErrorHandler {
	return WithStreamServerFailHandler(func(c context.Context, req interface{}, resp interface{}, info *grpc.StreamServerInfo, err *fail.Error) error {
		if err.Report {
			return f(c, req, resp, info, err)
		}
		return err
	})
}
