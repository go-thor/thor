package errors

import (
	"errors"
	"fmt"
)

// Error codes
const (
	ErrorCodeUnknown          = "unknown"
	ErrorCodeInvalidArgument  = "invalid_argument"
	ErrorCodeNotFound         = "not_found"
	ErrorCodeAlreadyExists    = "already_exists"
	ErrorCodePermissionDenied = "permission_denied"
	ErrorCodeUnauthenticated  = "unauthenticated"
	ErrorCodeTimeout          = "timeout"
	ErrorCodeCancelled        = "cancelled"
	ErrorCodeDeadlineExceeded = "deadline_exceeded"
)

// Common errors
var (
	ErrServerClosed     = New(ErrorCodeUnknown, "server closed")
	ErrTimeout          = New(ErrorCodeTimeout, "timeout")
	ErrCancelled        = New(ErrorCodeCancelled, "cancelled")
	ErrDeadlineExceeded = New(ErrorCodeDeadlineExceeded, "deadline exceeded")
	ErrClientClosed     = New(ErrorCodeUnknown, "client closed")
	ErrServiceNotFound  = New(ErrorCodeNotFound, "service not found")
	ErrMethodNotFound   = New(ErrorCodeNotFound, "method not found")
)

// Error represents an error with additional context
type Error struct {
	Code    string
	Message string
	Cause   error
}

// Error returns the error message
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the cause of the error
func (e *Error) Unwrap() error {
	return e.Cause
}

// New creates a new error
func New(code, message string) error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Newf creates a new error with formatted message
func Newf(code, format string, args ...interface{}) error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// Wrap wraps an error with additional context
func Wrap(code string, err error, message string) error {
	if err == nil {
		return nil
	}
	return &Error{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// Is reports whether any error in err's chain matches target
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Code returns the error code
func Code(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return ErrorCodeUnknown
}

// Message returns the error message
func Message(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return e.Message
	}
	return err.Error()
}

// Cause returns the cause of the error
func Cause(err error) error {
	var e *Error
	if errors.As(err, &e) {
		return e.Cause
	}
	return nil
}
