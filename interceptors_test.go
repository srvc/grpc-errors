package grpcerrors

import (
	"errors"
	"reflect"
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/srvc/fail"
	"github.com/srvc/grpc-errors/testing"
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

type failService struct {
}

func (s *failService) EmptyCall(context.Context, *errorstesting.Empty) (*errorstesting.Empty, error) {
	return nil, fail.New("This error is wrapped with fail.Error")
}

type ignoredErrorService struct {
}

func (s *ignoredErrorService) EmptyCall(context.Context, *errorstesting.Empty) (*errorstesting.Empty, error) {
	return nil, fail.Wrap(errors.New("This error should be ignored"), fail.WithIgnorable())
}

type errorWithStatusService struct {
	Code int
}

func (s *errorWithStatusService) EmptyCall(context.Context, *errorstesting.Empty) (*errorstesting.Empty, error) {
	return nil, fail.Wrap(errors.New("This error has a status code"), fail.WithCode(s.Code))
}

type errorWithGrpcStatusService struct {
	Code codes.Code
}

func (s *errorWithGrpcStatusService) EmptyCall(context.Context, *errorstesting.Empty) (*errorstesting.Empty, error) {
	return nil, fail.Wrap(status.Error(s.Code, "This error has a gRPC status code"))
}

// Testings
// ================================================
func Test_UnaryServerInterceptor(t *testing.T) {
	cases := []struct {
		test                  string
		server                errorstesting.TestServiceServer
		code                  codes.Code
		errored               bool
		notWrapped            bool
		reportable            bool
		handleNotWrappedError func(err error) error
	}{
		{
			test:   "no errors",
			server: &emptyService{},
			code:   codes.OK,
		},
		{
			test:       "unwrapped error",
			server:     &errorService{},
			code:       codes.Unknown,
			errored:    true,
			notWrapped: true,
			reportable: false,
		},
		{
			test:                  "wrap unwrapped error",
			server:                &errorService{},
			code:                  codes.Unknown,
			errored:               true,
			notWrapped:            true,
			reportable:            true,
			handleNotWrappedError: func(err error) error { return fail.Wrap(err) },
		},
		{
			test:       "wrapped error",
			server:     &failService{},
			code:       codes.Unknown,
			errored:    true,
			notWrapped: false,
			reportable: true,
		},
		{
			test:       "ignored error",
			server:     &ignoredErrorService{},
			code:       codes.Unknown,
			errored:    true,
			notWrapped: false,
			reportable: false,
		},
		{
			test:       "error with code that contained CodeMap",
			server:     &errorWithStatusService{Code: 50},
			code:       codes.PermissionDenied,
			errored:    true,
			notWrapped: false,
			reportable: true,
		},
		{
			test:       "error with unknown code",
			server:     &errorWithStatusService{Code: 51},
			code:       codes.Unknown,
			errored:    true,
			notWrapped: false,
			reportable: true,
		},
		{
			test:       "error with code that can handle with CodeMapper",
			server:     &errorWithStatusService{Code: 52},
			code:       codes.InvalidArgument,
			errored:    true,
			notWrapped: false,
			reportable: true,
		},
		{
			test:       "error with gRPC's code",
			server:     &errorWithGrpcStatusService{Code: codes.AlreadyExists},
			code:       codes.AlreadyExists,
			errored:    true,
			notWrapped: false,
			reportable: true,
		},
	}

	for _, c := range cases {
		t.Run(c.test, func(t *testing.T) {
			var notWrapped, reportable bool

			ctx := errorstesting.CreateTestContext(t)
			ctx.Service = c.server
			ctx.AddUnaryServerInterceptor(
				UnaryServerInterceptor(
					WithNotWrappedErrorHandler(func(_ context.Context, err error) error {
						notWrapped = true
						if c.handleNotWrappedError != nil {
							return c.handleNotWrappedError(err)
						}
						return err
					}),
					WithReportableErrorHandler(func(_ context.Context, err *fail.Error) error {
						reportable = true
						return err
					}),
					WithGrpcStatusUnwrapper(),
					WithStatusCodeMap(CodeMap{50: codes.PermissionDenied}),
					WithStatusCodeMapper(func(c interface{}) codes.Code {
						if c == 52 {
							return codes.InvalidArgument
						}
						return codes.Unknown
					}),
				),
			)
			ctx.Setup()
			defer ctx.Teardown()

			resp, err := ctx.Client.EmptyCall(context.Background(), &errorstesting.Empty{})

			if c.errored {
				if resp != nil {
					t.Error("The request should not return a response")
				}

				if err == nil {
					t.Error("The request should return an error")
				}
			} else {
				if resp == nil {
					t.Error("The request should return a response")
				}

				if err != nil {
					t.Error("The request should not return any errors")
				}
			}

			if got, want := notWrapped, c.notWrapped; got != want {
				t.Errorf("The returned error is wrapped: got %t, want %t", got, want)
			}

			if got, want := reportable, c.reportable; got != want {
				t.Errorf("The returned error is reportable: got %t, want %t", got, want)
			}

			if st, ok := status.FromError(err); ok {
				if got, want := st.Code(), c.code; got != want {
					t.Errorf("The returned error has error code %v, want %v", got, want)
				}
			} else {
				t.Errorf("The returned error does not have error code: %v", err)
			}
		})
	}
}

// Testings for UnaryServerErrorHandler
// ================================================
func Test_UnaryServerInterceptor_WithUnaryServerReportableErrorHandler_WhenAnErrorIsNotAnnotatedWithReport(t *testing.T) {
	called := false
	req := &errorstesting.Empty{}

	ctx := errorstesting.CreateTestContext(t)
	ctx.Service = &failService{}
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

	if !called {
		t.Error("Report error handler should not be called")
	}
}

func Test_UnaryServerInterceptor_WithUnaryServerReportableErrorHandler_WhenAnErrorIsAnnotatedWithIgnored(t *testing.T) {
	called := false
	req := &errorstesting.Empty{}

	ctx := errorstesting.CreateTestContext(t)
	ctx.Service = &ignoredErrorService{}
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

	if called {
		t.Error("Report error handler should be called")
	}
}
