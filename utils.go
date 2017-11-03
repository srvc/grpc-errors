package grpcerrors

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func composeHandlers(funcs []ErrorHandlerFunc) ErrorHandlerFunc {
	return func(c context.Context, err error) error {
		if err != nil {
			for _, f := range funcs {
				err = f(c, err)
				if err == nil {
					break
				}
			}
		}
		return err
	}
}

type composedUnaryServerErrorHandler struct {
	handlers []UnaryServerErrorHandler
}

func composeUnaryServerErrorHandlers(handlers []UnaryServerErrorHandler) UnaryServerErrorHandler {
	return &composedUnaryServerErrorHandler{
		handlers: handlers,
	}
}

func (ch *composedUnaryServerErrorHandler) HandleUnaryServerError(
	c context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	err error,
) error {
	if err != nil {
		for _, h := range ch.handlers {
			err = h.HandleUnaryServerError(c, req, info, err)
			if err == nil {
				break
			}
		}
	}
	return err
}

type composedStreamServerErrorHandler struct {
	handlers []StreamServerErrorHandler
}

func composeStreamServerErrorHandlers(handlers []StreamServerErrorHandler) StreamServerErrorHandler {
	return &composedStreamServerErrorHandler{
		handlers: handlers,
	}
}

func (ch *composedStreamServerErrorHandler) HandleStreamServerError(
	c context.Context,
	req interface{},
	resp interface{},
	info *grpc.StreamServerInfo,
	err error,
) error {
	if err != nil {
		for _, h := range ch.handlers {
			err = h.HandleStreamServerError(c, req, resp, info, err)
			if err == nil {
				break
			}
		}
	}
	return err
}
