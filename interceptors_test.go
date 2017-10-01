package grpcerrors

import (
	"context"
	"errors"
	"testing"

	"github.com/creasty/apperrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/izumin5210/grpc-errors/testing"
)

// Sevice implementations
// ================================================
type emptyService struct {
}

func (s *emptyService) EmptyCall(context.Context, *errorstesting.Empty) (*errorstesting.Empty, error) {
	return &errorstesting.Empty{}, nil
}

type errorService struct {
}

func (s *errorService) EmptyCall(context.Context, *errorstesting.Empty) (*errorstesting.Empty, error) {
	return nil, errors.New("This error is not wrapped with apperrors.Error")
}

type appErrorService struct {
}

func (s *appErrorService) EmptyCall(context.Context, *errorstesting.Empty) (*errorstesting.Empty, error) {
	return nil, apperrors.New("This error is wrapped with apperrors.Error")
}

type reportErrorService struct {
}

func (s *reportErrorService) EmptyCall(context.Context, *errorstesting.Empty) (*errorstesting.Empty, error) {
	return nil, apperrors.WithReport(apperrors.New("This error should be reported"))
}

type errorWithStatusService struct {
	Code int
}

func (s *errorWithStatusService) EmptyCall(context.Context, *errorstesting.Empty) (*errorstesting.Empty, error) {
	return nil, apperrors.WithStatusCode(errors.New("This error has a status code"), s.Code)
}

// Testings
// ================================================
func Test_UnaryServerInterceptor_WhenDoesNotRespondErrors(t *testing.T) {
	called := false

	ctx := errorstesting.CreateTestContext(t)
	ctx.Service = &emptyService{}
	ctx.AddUnaryServerInterceptor(
		UnaryServerInterceptor(
			HandleNotWrappedError(func(err error) error {
				called = true
				return err
			}),
			Report(func(err *apperrors.Error) error {
				called = true
				return err
			}),
		),
	)
	ctx.Setup()
	defer ctx.Teardown()

	resp, err := ctx.Client.EmptyCall(context.Background(), &errorstesting.Empty{})

	if resp == nil {
		t.Error("The request should return a response")
	}

	if err != nil {
		t.Error("The request should not return any errors")
	}

	if called {
		t.Error("Report error handler should not be called")
	}
}

func Test_UnaryServerInterceptor_HandleNotWrappedError_WhenAnErrorIsNotWrappedWithAppError(t *testing.T) {
	called := false

	ctx := errorstesting.CreateTestContext(t)
	ctx.Service = &errorService{}
	ctx.AddUnaryServerInterceptor(
		UnaryServerInterceptor(
			HandleNotWrappedError(func(err error) error {
				called = true
				return err
			}),
		),
	)
	ctx.Setup()
	defer ctx.Teardown()

	resp, err := ctx.Client.EmptyCall(context.Background(), &errorstesting.Empty{})

	if resp != nil {
		t.Error("The request should not return any responses")
	}

	if err == nil {
		t.Error("The request should return an error")
	}

	if !called {
		t.Error("Report error handler should be called")
	}
}

func Test_UnaryServerInterceptor_HandleNotWrappedError_WhenAnErrorIsWrappedWithAppError(t *testing.T) {
	called := false

	ctx := errorstesting.CreateTestContext(t)
	ctx.Service = &appErrorService{}
	ctx.AddUnaryServerInterceptor(
		UnaryServerInterceptor(
			HandleNotWrappedError(func(err error) error {
				called = true
				return err
			}),
		),
	)
	ctx.Setup()
	defer ctx.Teardown()

	resp, err := ctx.Client.EmptyCall(context.Background(), &errorstesting.Empty{})

	if resp != nil {
		t.Error("The request should not return any responses")
	}

	if err == nil {
		t.Error("The request should return an error")
	}

	if called {
		t.Error("Report error handler should not be called")
	}
}

func Test_UnaryServerInterceptor_Reporting_WhenAnErrorIsNotAnnotatedWithReport(t *testing.T) {
	called := false

	ctx := errorstesting.CreateTestContext(t)
	ctx.Service = &appErrorService{}
	ctx.AddUnaryServerInterceptor(
		UnaryServerInterceptor(
			Report(func(err *apperrors.Error) error {
				called = true
				return err
			}),
		),
	)
	ctx.Setup()
	defer ctx.Teardown()

	resp, err := ctx.Client.EmptyCall(context.Background(), &errorstesting.Empty{})

	if resp != nil {
		t.Error("The request should not return any responses")
	}

	if err == nil {
		t.Error("The request should return an error")
	}

	if called {
		t.Error("Report error handler should not be called")
	}
}

func Test_UnaryServerInterceptor_Reporting_WhenAnErrorIsAnnotatedWithReport(t *testing.T) {
	called := false

	ctx := errorstesting.CreateTestContext(t)
	ctx.Service = &reportErrorService{}
	ctx.AddUnaryServerInterceptor(
		UnaryServerInterceptor(
			Report(func(err *apperrors.Error) error {
				called = true
				return err
			}),
		),
	)
	ctx.Setup()
	defer ctx.Teardown()

	resp, err := ctx.Client.EmptyCall(context.Background(), &errorstesting.Empty{})

	if resp != nil {
		t.Error("The request should not return any responses")
	}

	if err == nil {
		t.Error("The request should return an error")
	}

	if !called {
		t.Error("Report error handler should be called")
	}
}

func Test_UnaryServerInterceptor_MapStatusCode(t *testing.T) {
	code := 50
	mappedCode := codes.Unavailable

	ctx := errorstesting.CreateTestContext(t)
	ctx.Service = &errorWithStatusService{Code: code}
	ctx.AddUnaryServerInterceptor(
		UnaryServerInterceptor(
			MapStatusCode(map[int]codes.Code{
				code: mappedCode,
			}),
		),
	)
	ctx.Setup()
	defer ctx.Teardown()

	resp, err := ctx.Client.EmptyCall(context.Background(), &errorstesting.Empty{})

	if resp != nil {
		t.Error("The request should not return any responses")
	}

	if err == nil {
		t.Error("The request should return an error")
	}

	if st, ok := status.FromError(err); !ok {
		t.Error("Returned error should has status code")
	} else if got, want := st.Code(), mappedCode; got != want {
		t.Errorf("Returned error had status code %v, want %v", got, want)
	}
}

func Test_UnaryServerInterceptor_MapStatusCode_WhenUnknownCode(t *testing.T) {
	code := 50
	mappedCode := codes.Unavailable

	ctx := errorstesting.CreateTestContext(t)
	ctx.Service = &errorWithStatusService{Code: code + 1}
	ctx.AddUnaryServerInterceptor(
		UnaryServerInterceptor(
			MapStatusCode(map[int]codes.Code{
				code: mappedCode,
			}),
		),
	)
	ctx.Setup()
	defer ctx.Teardown()

	resp, err := ctx.Client.EmptyCall(context.Background(), &errorstesting.Empty{})

	if resp != nil {
		t.Error("The request should not return any responses")
	}

	if err == nil {
		t.Error("The request should return an error")
	}

	if st, ok := status.FromError(err); !ok {
		t.Error("Returned error should has status code")
	} else if got, want := st.Code(), codes.Internal; got != want {
		t.Errorf("Returned error had status code %v, want %v", got, want)
	}
}
