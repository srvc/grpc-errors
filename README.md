# grpc-errors
[![Build Status](https://travis-ci.org/izumin5210/grpc-errors.svg?branch=master)](https://travis-ci.org/izumin5210/grpc-errors)
[![codecov](https://codecov.io/gh/izumin5210/grpc-errors/branch/master/graph/badge.svg)](https://codecov.io/gh/izumin5210/grpc-errors)
[![GoDoc](https://godoc.org/github.com/izumin5210/grpc-errors?status.svg)](https://godoc.org/github.com/izumin5210/grpc-errors)
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

	s := grpc.NewServer(
		grpc_middleware.WithStreamServerChain(
			grpcerrors.StreamServerInterceptor(
				grpcerrors.WithNotWrappedErrorHandler(func(err error) error {
					return apperrors.WithStatusCode(err, CodeNotWrapped)
				}),
				grpcerrors.WithReportableErrorHandler(func(err *apperrors.Error) error {
					switch err.StatusCode {
					case CodeYourCustomError:
						// Report your custom errors
					case CodeNotWrapped:
						// Report not wrapped errors
					default:
						// Report errors
					}
					return err
				}),
				grpcerrors.WithStatusCodeMapper(grpcCodeByYourCode),
			),
		),
		grpc_middleware.WithUnaryServerChain(
			grpcerrors.UnaryServerInterceptor(
				// Write your error handlers for an unary server
			),
		),
	)

	// Register server implementations

	s.Serve(lis)
}
```
