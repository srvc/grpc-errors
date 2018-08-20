# grpc-errors
[![Build Status](https://travis-ci.org/srvc/grpc-errors.svg?branch=master)](https://travis-ci.org/srvc/grpc-errors)
[![codecov](https://codecov.io/gh/srvc/grpc-errors/branch/master/graph/badge.svg)](https://codecov.io/gh/srvc/grpc-errors)
[![GoDoc](https://godoc.org/github.com/srvc/grpc-errors?status.svg)](https://godoc.org/github.com/srvc/grpc-errors)
[![Go project version](https://badge.fury.io/go/github.com%2Fizumin5210%2Fgrpc-errors.svg)](https://badge.fury.io/go/github.com%2Fizumin5210%2Fgrpc-errors)
[![Go Report Card](https://goreportcard.com/badge/github.com/srvc/grpc-errors)](https://goreportcard.com/report/github.com/srvc/grpc-errors)
[![license](https://img.shields.io/github/license/srvc/grpc-errors.svg)](./LICENSE)

`grpc-errors` is a middleware providing better error handling to resolve errors easily.

## Example

```go
package main

import (
	"context"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/srvc/fail"
	"github.com/srvc/grpc-errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const (
	CodeOK uint32 = iota
	CodeInvalidArgument
	CodeNotFound
	CodeYourCustomError
	CodeNotWrapped
	CodeUnknown
)

var grpcCodeByYourCode = grpcerrors.CodeMap{
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
			// WithNotWrappedErrorHandler handles an error not wrapped with `*fail.Error`.
			// A handler function should wrap received error with `*fail.Error`.
			return fail.Wrap(err, fail.WithCode(CodeNotWrapped))
		}),
		grpcerrors.WithReportableErrorHandler(func(c context.Context, err *fail.Error) error {
			// WithReportableErrorHandler handles an erorr annotated with the reportability.
			// You reports to an external service if necessary.
			// And you can attach request contexts to error reports.
			return err
		}),
		grpcerrors.WithCodeMap(grpcCodeByYourCode),
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
