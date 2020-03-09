## Unreleased

## 1.2.0

- [Technically breaking] Use github.com/srvc/fail v4.1.1 https://github.com/srvc/grpc-errors/pull/21

## 1.1.0

- Use go mod instead of dep https://github.com/srvc/grpc-errors/pull/20

## 1.0.0

- [Breaking] Expand error handler interfaces https://github.com/srvc/grpc-errors/pull/11
- Add new error handler for catching reportable errors for unary server https://github.com/srvc/grpc-errors/pull/12
- [Behavioral change] Use codes.Unknown on WithStatusCodeMap when a status code is unknown https://github.com/srvc/grpc-errors/pull/14
- [Behavioral change] Map status code only when grpc/codes.Code has not been set https://github.com/srvc/grpc-errors/pull/15
- [Behavioral change] WithStatusCodeMap: return original error when no code is found https://github.com/srvc/grpc-errors/pull/16
- Add `WithGrpcStatusUnwrapper` https://github.com/srvc/grpc-errors/pull/16
- Add error handler for stream servers https://github.com/srvc/grpc-errors/pull/17
- Move from izumin5210/grpc-errors to srvc/grpc-errors https://github.com/srvc/grpc-errors/pull/19
- [Breaking] Use srvc/fail instead of creasty/apperrors https://github.com/srvc/grpc-errors/pull/18


## 0.2.0

- Add `WithAppErrorHandler` https://github.com/srvc/grpc-errors/pull/7
- [Breaking] Make error handlers receive request contexts https://github.com/srvc/grpc-errors/pull/8
- [Breaking] Add new status code mapper handler impl https://github.com/srvc/grpc-errors/pull/9
- Improve readme https://github.com/srvc/grpc-errors/pull/10

## 0.1.0

Initial release.
