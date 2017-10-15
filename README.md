# grpc-errors
[![Build Status](https://travis-ci.org/izumin5210/grpc-errors.svg?branch=master)](https://travis-ci.org/izumin5210/grpc-errors)
[![codecov](https://codecov.io/gh/izumin5210/grpc-errors/branch/master/graph/badge.svg)](https://codecov.io/gh/izumin5210/grpc-errors)
[![GoDoc](https://godoc.org/github.com/izumin5210/grpc-errors?status.svg)](https://godoc.org/github.com/izumin5210/grpc-errors)
[![Go project version](https://badge.fury.io/go/github.com%2Fizumin5210%2Fgrpc-errors.svg)](https://badge.fury.io/go/github.com%2Fizumin5210%2Fgrpc-errors)
[![Go Report Card](https://goreportcard.com/badge/github.com/izumin5210/grpc-errors)](https://goreportcard.com/report/github.com/izumin5210/grpc-errors)
[![license](https://img.shields.io/github/license/izumin5210/grpc-errors.svg)](./LICENSE)

`grpc-errors` is a middleware providing better error handling to resolve errors easily.

## Example

```go
package main

import (
	"context"
	"net"

	"github.com/creasty/apperrors"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/izumin5210/grpc-errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const (
	CodeOK int = iota
	CodeInvalidArgument
	CodeNotFound
	CodeYourCustomError
	CodeNotWrapped
	CodeUnknown
)

var grpcCodeByYourCode = map[int]codes.Code{
	CodeOK:              codes.OK,
	CodeInvalidArgument: codes.InvalidArgument,
	CodeNotFound:        codes.NotFound,
}

func main() {
	lis, err := net.Listen("tcp", "api.example.com:80")
	if err != nil {
		panic(err)
	}

	errorHandlers := []grpcerrors.ErrorHandlerFunc{
		grpcerrors.WithNotWrappedErrorHandler(func(c context.Context, err error) error {
			// WithNotWrappedErrorHandler handles an error not wrapped with `*apperror.Error`.
			// A handler function should wrap received error with `*apperror.Error`.
			return apperrors.WithStatusCode(err, CodeNotWrapped)
		}),
		grpcerrors.WithReportableErrorHandler(func(c context.Context, err *apperrors.Error) error {
			// WithReportableErrorHandler handles an erorr annotated with the reportability.
			// You reports to an external service if necessary.
			// And you can attach request contexts to error reports.
			return err
		}),
		grpcerrors.WithStatusCodeMap(grpcCodeByYourCode),
	}

	s := grpc.NewServer(
		grpc_middleware.WithStreamServerChain(
			grpcerrors.StreamServerInterceptor(errorHandlers...),
		),
		grpc_middleware.WithUnaryServerChain(
			grpcerrors.UnaryServerInterceptor(errorHandlers...),
		),
	)

	// Register server implementations

	s.Serve(lis)
}
```
