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
	"github.com/izumin5210/grpc-errors"
	"github.com/grpc-ecosystem/go-grpc-middleware"
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
	lis, err := net.Listen("tcp", ctx.Config.Host)
	if err != nil {
		return err
	}

	s := grpc.NewServer(
		grpc_middleware.WithStreamServerChain(
			grpcerrors.StreamServerInterceptor(
				grpcerrors.HandleNotWrappedError(func(err error) error {
					return apperrors.WithStatusCode(err, CodeNotWrapped)
				}),
				grpcerrors.Report(func(err *apperrors.Error) error {
					swtich err.StatusCode {
					case CodeYourCustomError:
						// Report your custom errors
					case CodeNotWrapped:
						// Report not wrapped errors
					default:
						// Report errors
					}
					return err
				}),
				grpcerrors.MapStatusCode(grpcCodeByYourCode),
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
