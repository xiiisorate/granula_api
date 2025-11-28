// Package errors provides domain-specific error types for Granula microservices.
//
// Errors include:
// - Standard error codes (NOT_FOUND, VALIDATION_ERROR, etc.)
// - gRPC status code mapping
// - Structured error details
// - Stack traces (in development mode)
//
// Example:
//
//	return errors.NotFound("user", userID)
//	return errors.Validation("email", "invalid format")
//	return errors.Internal("database connection failed").WithCause(err)
package errors

import (
	"fmt"
	"runtime"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Code represents an error code.
type Code string

// Standard error codes.
const (
	CodeUnknown           Code = "UNKNOWN"
	CodeInternal          Code = "INTERNAL"
	CodeNotFound          Code = "NOT_FOUND"
	CodeAlreadyExists     Code = "ALREADY_EXISTS"
	CodeInvalidArgument   Code = "INVALID_ARGUMENT"
	CodeValidation        Code = "VALIDATION_ERROR"
	CodeUnauthenticated   Code = "UNAUTHENTICATED"
	CodeUnauthorized      Code = "UNAUTHORIZED"
	CodePermissionDenied  Code = "PERMISSION_DENIED"
	CodeRateLimited       Code = "RATE_LIMITED"
	CodeTimeout           Code = "TIMEOUT"
	CodeCancelled         Code = "CANCELLED"
	CodeConflict          Code = "CONFLICT"
	CodePrecondition      Code = "PRECONDITION_FAILED"
	CodeUnavailable       Code = "UNAVAILABLE"
	CodeResourceExhausted Code = "RESOURCE_EXHAUSTED"
)

// Error represents a domain error.
type Error struct {
	// Code is the error code.
	Code Code `json:"code"`

	// Message is a human-readable message.
	Message string `json:"message"`

	// Details contains additional error information.
	Details map[string]string `json:"details,omitempty"`

	// Cause is the underlying error.
	Cause error `json:"-"`

	// Stack is the stack trace (if enabled).
	Stack string `json:"-"`

	// GRPCCode is the corresponding gRPC status code.
	GRPCCode codes.Code `json:"-"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause.
func (e *Error) Unwrap() error {
	return e.Cause
}

// WithCause sets the underlying cause.
func (e *Error) WithCause(cause error) *Error {
	e.Cause = cause
	return e
}

// WithDetail adds a detail to the error.
func (e *Error) WithDetail(key, value string) *Error {
	if e.Details == nil {
		e.Details = make(map[string]string)
	}
	e.Details[key] = value
	return e
}

// WithDetails adds multiple details.
func (e *Error) WithDetails(details map[string]string) *Error {
	if e.Details == nil {
		e.Details = make(map[string]string)
	}
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// WithStack captures the current stack trace.
func (e *Error) WithStack() *Error {
	e.Stack = captureStack(3)
	return e
}

// ToGRPCStatus converts the error to a gRPC status.
func (e *Error) ToGRPCStatus() *status.Status {
	return status.New(e.GRPCCode, e.Message)
}

// ToGRPCError converts the error to a gRPC error.
func (e *Error) ToGRPCError() error {
	return e.ToGRPCStatus().Err()
}

// Is checks if the target error has the same code.
func (e *Error) Is(target error) bool {
	if t, ok := target.(*Error); ok {
		return e.Code == t.Code
	}
	return false
}

// captureStack captures the current stack trace.
func captureStack(skip int) string {
	var sb strings.Builder
	pcs := make([]uintptr, 32)
	n := runtime.Callers(skip, pcs)
	frames := runtime.CallersFrames(pcs[:n])

	for {
		frame, more := frames.Next()
		// Skip runtime and testing frames
		if strings.Contains(frame.File, "runtime/") ||
			strings.Contains(frame.File, "testing/") {
			if !more {
				break
			}
			continue
		}
		fmt.Fprintf(&sb, "%s:%d %s\n", frame.File, frame.Line, frame.Function)
		if !more {
			break
		}
	}
	return sb.String()
}

// newError creates a new Error with default values.
func newError(code Code, grpcCode codes.Code, message string) *Error {
	return &Error{
		Code:     code,
		Message:  message,
		GRPCCode: grpcCode,
	}
}

// -----------------------------------------------------------------------------
// Error constructors
// -----------------------------------------------------------------------------

// NotFound creates a NOT_FOUND error.
func NotFound(resource, id string) *Error {
	return newError(CodeNotFound, codes.NotFound,
		fmt.Sprintf("%s with id '%s' not found", resource, id))
}

// NotFoundMsg creates a NOT_FOUND error with custom message.
func NotFoundMsg(message string) *Error {
	return newError(CodeNotFound, codes.NotFound, message)
}

// AlreadyExists creates an ALREADY_EXISTS error.
func AlreadyExists(resource, field, value string) *Error {
	return newError(CodeAlreadyExists, codes.AlreadyExists,
		fmt.Sprintf("%s with %s '%s' already exists", resource, field, value)).
		WithDetail("field", field).
		WithDetail("value", value)
}

// InvalidArgument creates an INVALID_ARGUMENT error.
func InvalidArgument(field, reason string) *Error {
	return newError(CodeInvalidArgument, codes.InvalidArgument,
		fmt.Sprintf("invalid argument '%s': %s", field, reason)).
		WithDetail("field", field)
}

// Validation creates a VALIDATION_ERROR.
func Validation(field, message string) *Error {
	return newError(CodeValidation, codes.InvalidArgument,
		fmt.Sprintf("validation error: %s - %s", field, message)).
		WithDetail("field", field)
}

// ValidationMultiple creates a VALIDATION_ERROR with multiple fields.
func ValidationMultiple(errors map[string]string) *Error {
	var messages []string
	for field, msg := range errors {
		messages = append(messages, fmt.Sprintf("%s: %s", field, msg))
	}
	return newError(CodeValidation, codes.InvalidArgument,
		"validation errors: "+strings.Join(messages, "; ")).
		WithDetails(errors)
}

// Unauthenticated creates an UNAUTHENTICATED error.
func Unauthenticated(message string) *Error {
	if message == "" {
		message = "authentication required"
	}
	return newError(CodeUnauthenticated, codes.Unauthenticated, message)
}

// Unauthorized creates an UNAUTHORIZED error.
func Unauthorized(action, resource string) *Error {
	return newError(CodeUnauthorized, codes.PermissionDenied,
		fmt.Sprintf("not authorized to %s %s", action, resource))
}

// PermissionDenied creates a PERMISSION_DENIED error.
func PermissionDenied(message string) *Error {
	if message == "" {
		message = "permission denied"
	}
	return newError(CodePermissionDenied, codes.PermissionDenied, message)
}

// Internal creates an INTERNAL error.
func Internal(message string) *Error {
	return newError(CodeInternal, codes.Internal, message)
}

// Internalf creates an INTERNAL error with formatted message.
func Internalf(format string, args ...any) *Error {
	return Internal(fmt.Sprintf(format, args...))
}

// RateLimited creates a RATE_LIMITED error.
func RateLimited(message string) *Error {
	if message == "" {
		message = "rate limit exceeded"
	}
	return newError(CodeRateLimited, codes.ResourceExhausted, message)
}

// Timeout creates a TIMEOUT error.
func Timeout(operation string) *Error {
	return newError(CodeTimeout, codes.DeadlineExceeded,
		fmt.Sprintf("operation '%s' timed out", operation))
}

// Cancelled creates a CANCELLED error.
func Cancelled(message string) *Error {
	if message == "" {
		message = "operation cancelled"
	}
	return newError(CodeCancelled, codes.Canceled, message)
}

// Conflict creates a CONFLICT error.
func Conflict(message string) *Error {
	return newError(CodeConflict, codes.Aborted, message)
}

// PreconditionFailed creates a PRECONDITION_FAILED error.
func PreconditionFailed(message string) *Error {
	return newError(CodePrecondition, codes.FailedPrecondition, message)
}

// Unavailable creates an UNAVAILABLE error.
func Unavailable(service string) *Error {
	return newError(CodeUnavailable, codes.Unavailable,
		fmt.Sprintf("service '%s' is unavailable", service))
}

// ResourceExhausted creates a RESOURCE_EXHAUSTED error.
func ResourceExhausted(resource string) *Error {
	return newError(CodeResourceExhausted, codes.ResourceExhausted,
		fmt.Sprintf("resource '%s' exhausted", resource))
}

// Wrap wraps an existing error with additional context.
func Wrap(err error, message string) *Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Error); ok {
		return &Error{
			Code:     e.Code,
			Message:  message + ": " + e.Message,
			Details:  e.Details,
			Cause:    e.Cause,
			GRPCCode: e.GRPCCode,
		}
	}
	return Internal(message).WithCause(err)
}

// Wrapf wraps an error with formatted message.
func Wrapf(err error, format string, args ...any) *Error {
	return Wrap(err, fmt.Sprintf(format, args...))
}

// -----------------------------------------------------------------------------
// Error checking
// -----------------------------------------------------------------------------

// IsNotFound checks if error is NOT_FOUND.
func IsNotFound(err error) bool {
	return hasCode(err, CodeNotFound)
}

// IsAlreadyExists checks if error is ALREADY_EXISTS.
func IsAlreadyExists(err error) bool {
	return hasCode(err, CodeAlreadyExists)
}

// IsValidation checks if error is VALIDATION_ERROR.
func IsValidation(err error) bool {
	return hasCode(err, CodeValidation)
}

// IsUnauthenticated checks if error is UNAUTHENTICATED.
func IsUnauthenticated(err error) bool {
	return hasCode(err, CodeUnauthenticated)
}

// IsUnauthorized checks if error is UNAUTHORIZED or PERMISSION_DENIED.
func IsUnauthorized(err error) bool {
	return hasCode(err, CodeUnauthorized) || hasCode(err, CodePermissionDenied)
}

// IsInternal checks if error is INTERNAL.
func IsInternal(err error) bool {
	return hasCode(err, CodeInternal)
}

// IsRateLimited checks if error is RATE_LIMITED.
func IsRateLimited(err error) bool {
	return hasCode(err, CodeRateLimited)
}

// hasCode checks if error has specific code.
func hasCode(err error, code Code) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == code
	}
	return false
}

// FromGRPCError converts gRPC error to Error.
func FromGRPCError(err error) *Error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return Internal(err.Error()).WithCause(err)
	}

	code := grpcCodeToCode(st.Code())
	return &Error{
		Code:     code,
		Message:  st.Message(),
		GRPCCode: st.Code(),
	}
}

// grpcCodeToCode maps gRPC codes to domain codes.
func grpcCodeToCode(c codes.Code) Code {
	switch c {
	case codes.NotFound:
		return CodeNotFound
	case codes.AlreadyExists:
		return CodeAlreadyExists
	case codes.InvalidArgument:
		return CodeInvalidArgument
	case codes.Unauthenticated:
		return CodeUnauthenticated
	case codes.PermissionDenied:
		return CodePermissionDenied
	case codes.ResourceExhausted:
		return CodeRateLimited
	case codes.DeadlineExceeded:
		return CodeTimeout
	case codes.Canceled:
		return CodeCancelled
	case codes.Aborted:
		return CodeConflict
	case codes.FailedPrecondition:
		return CodePrecondition
	case codes.Unavailable:
		return CodeUnavailable
	default:
		return CodeInternal
	}
}
