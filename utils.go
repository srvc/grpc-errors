package grpcerrors

import (
	"golang.org/x/net/context"
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
