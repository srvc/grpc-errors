package grpcerrors

import (
	"errors"
	"reflect"
	"testing"

	"github.com/izumin5210/grpc-errors/testing"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/izumin5210/fail"
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
	return nil, errors.New("This error is not wrapped with fail.Error")
}

type appErrorService struct {
}

func (s *appErrorService) EmptyCall(context.Context, *errorstesting.Empty) (*errorstesting.Empty, error) {
	return nil, fail.New("This error is wrapped with fail.Error")
}

type reportErrorService struct {
}

func (s *reportErrorService) EmptyCall(context.Context, *errorstesting.Empty) (*errorstesting.Empty, error) {
	return nil, fail.Wrap(fail.New("This error should be reported"), fail.WithReport())
}

type errorWithStatusService struct {
	Code int
}

func (s *errorWithStatusService) EmptyCall(context.Context, *errorstesting.Empty) (*errorstesting.Empty, error) {
	return nil, fail.Wrap(errors.New("This error has a status code"), fail.WithStatusCode(s.Code))
}

type errorWithGrpcStatusService struct {
	Code codes.Code
}

func (s *errorWithGrpcStatusService) EmptyCall(context.Context, *errorstesting.Empty) (*errorstesting.Empty, error) {
	return nil, fail.Wrap(status.Error(s.Code, "This error has a gRPC status code"))
}

// Testings
// ================================================
func Test_UnaryServerInterceptor_WhenDoesNotRespondErrors(t *testing.T) {
	called := false

	ctx := errorstesting.CreateTestContext(t)
	ctx.Service = &emptyService{}
	ctx.AddUnaryServerInterceptor(
		UnaryServerInterceptor(
			WithNotWrappedErrorHandler(func(_ context.Context, err error) error {
				called = true
				return err
			}),
			WithReportableErrorHandler(func(_ context.Context, err *fail.Error) error {
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

func Test_UnaryServerInterceptor_WithNotWrappedErrorHandler_WhenAnErrorIsNotWrappedWithAppError(t *testing.T) {
	called := false

	ctx := errorstesting.CreateTestContext(t)
	ctx.Service = &errorService{}
	ctx.AddUnaryServerInterceptor(
		UnaryServerInterceptor(
			WithNotWrappedErrorHandler(func(_ context.Context, err error) error {
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

func Test_UnaryServerInterceptor_WithNotWrappedErrorHandler_WhenAnErrorIsWrappedWithAppError(t *testing.T) {
	called := false

	ctx := errorstesting.CreateTestContext(t)
	ctx.Service = &appErrorService{}
	ctx.AddUnaryServerInterceptor(
		UnaryServerInterceptor(
			WithNotWrappedErrorHandler(func(_ context.Context, err error) error {
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

func Test_UnaryServerInterceptor_WithReportableErrorHandler_WhenAnErrorIsNotAnnotatedWithReport(t *testing.T) {
	called := false

	ctx := errorstesting.CreateTestContext(t)
	ctx.Service = &appErrorService{}
	ctx.AddUnaryServerInterceptor(
		UnaryServerInterceptor(
			WithReportableErrorHandler(func(_ context.Context, err *fail.Error) error {
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

func Test_UnaryServerInterceptor_WithReportableErrorHandler_WhenAnErrorIsAnnotatedWithReport(t *testing.T) {
	called := false

	ctx := errorstesting.CreateTestContext(t)
	ctx.Service = &reportErrorService{}
	ctx.AddUnaryServerInterceptor(
		UnaryServerInterceptor(
			WithReportableErrorHandler(func(_ context.Context, err *fail.Error) error {
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

func Test_UnaryServerInterceptor_WithStatusCodeMap(t *testing.T) {
	code := 50
	mappedCode := codes.Unavailable

	ctx := errorstesting.CreateTestContext(t)
	ctx.Service = &errorWithStatusService{Code: code}
	ctx.AddUnaryServerInterceptor(
		UnaryServerInterceptor(
			WithStatusCodeMap(map[int]codes.Code{
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

func Test_UnaryServerInterceptor_WithStatusCodeMap_WhenUnknownCode(t *testing.T) {
	code := 50
	mappedCode := codes.Unavailable

	ctx := errorstesting.CreateTestContext(t)
	ctx.Service = &errorWithStatusService{Code: code + 1}
	ctx.AddUnaryServerInterceptor(
		UnaryServerInterceptor(
			WithStatusCodeMap(map[int]codes.Code{
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
	} else if got, want := st.Code(), codes.Unknown; got != want {
		t.Errorf("Returned error had status code %v, want %v", got, want)
	}
}

func Test_UnaryServerInterceptor_WithGrpcStatusUnwrapper(t *testing.T) {
	code := codes.Unauthenticated

	ctx := errorstesting.CreateTestContext(t)
	ctx.Service = &errorWithGrpcStatusService{Code: code}
	ctx.AddUnaryServerInterceptor(
		UnaryServerInterceptor(
			WithGrpcStatusUnwrapper(),
			WithStatusCodeMap(map[int]codes.Code{
				50: codes.Unavailable,
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
	} else if got, want := st.Code(), code; got != want {
		t.Errorf("Returned error had status code %v, want %v", got, want)
	}
}

func Test_UnaryServerInterceptor_WithGrpcStatusUnwrapper_WithoutGrpcStatus(t *testing.T) {
	code := 50
	ctx := errorstesting.CreateTestContext(t)
	ctx.Service = &errorWithStatusService{Code: code}
	ctx.AddUnaryServerInterceptor(
		UnaryServerInterceptor(
			WithGrpcStatusUnwrapper(),
			WithStatusCodeMap(map[int]codes.Code{
				code: codes.Unauthenticated,
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
	} else if got, want := st.Code(), codes.Unauthenticated; got != want {
		t.Errorf("Returned error had status code %v, want %v", got, want)
	}
}

func Test_UnaryServerInterceptor_WithStatusCodeMapper(t *testing.T) {
	code := 50
	mappedCode := codes.Unavailable

	ctx := errorstesting.CreateTestContext(t)
	ctx.Service = &errorWithStatusService{Code: code}
	ctx.AddUnaryServerInterceptor(
		UnaryServerInterceptor(
			WithStatusCodeMapper(func(c int) codes.Code {
				if got, want := c, code; got != want {
					t.Errorf("Mapper func received %d, want %d", got, want)
				}
				return mappedCode
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

// Testings for UnaryServerErrorHandler
// ================================================
func Test_UnaryServerInterceptor_WithUnaryServerReportableErrorHandler_WhenAnErrorIsNotAnnotatedWithReport(t *testing.T) {
	called := false
	req := &errorstesting.Empty{}

	ctx := errorstesting.CreateTestContext(t)
	ctx.Service = &appErrorService{}
	ctx.AddUnaryServerInterceptor(
		UnaryServerInterceptor(
			WithUnaryServerReportableErrorHandler(func(_ context.Context, gotReq interface{}, info *grpc.UnaryServerInfo, err *fail.Error) error {
				called = true
				return err
			}),
		),
	)
	ctx.Setup()
	defer ctx.Teardown()

	resp, err := ctx.Client.EmptyCall(context.Background(), req)

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

func Test_UnaryServerInterceptor_WithUnaryServerReportableErrorHandler_WhenAnErrorIsAnnotatedWithReport(t *testing.T) {
	called := false
	req := &errorstesting.Empty{}

	ctx := errorstesting.CreateTestContext(t)
	ctx.Service = &reportErrorService{}
	ctx.AddUnaryServerInterceptor(
		UnaryServerInterceptor(
			WithUnaryServerReportableErrorHandler(func(_ context.Context, gotReq interface{}, info *grpc.UnaryServerInfo, err *fail.Error) error {
				called = true
				if got, want := gotReq, req; !reflect.DeepEqual(got, want) {
					t.Errorf("Received request is %v, want %v", got, want)
				}
				return err
			}),
		),
	)
	ctx.Setup()
	defer ctx.Teardown()

	resp, err := ctx.Client.EmptyCall(context.Background(), req)

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
