// Package apperr defines domain-agnostic application errors used across
// the platform. Domain packages should define their own sentinel errors
// that wrap these base types.
package apperr

import (
"errors"
"fmt"
"net/http"
)

// Kind categorizes errors for consistent HTTP status mapping.
type Kind int

const (
KindUnknown Kind = iota
KindNotFound
KindInvalid
KindConflict
KindUnauthorized
KindForbidden
KindInternal
KindUnavailable
KindRateLimited
)

// Error is a structured application error.
type Error struct {
Kind    Kind
Code    string // machine-readable code (e.g., "service_not_found")
Message string // human-readable message
Details map[string]string
err     error // wrapped error
}

func (e *Error) Error() string {
if e.err != nil {
return fmt.Sprintf("%s: %v", e.Message, e.err)
}
return e.Message
}

// Unwrap supports errors.Is and errors.As.
func (e *Error) Unwrap() error {
return e.err
}

// HTTPStatus maps an application error to an HTTP status code.
func (e *Error) HTTPStatus() int {
switch e.Kind {
case KindNotFound:
return http.StatusNotFound
case KindInvalid:
return http.StatusBadRequest
case KindConflict:
return http.StatusConflict
case KindUnauthorized:
return http.StatusUnauthorized
case KindForbidden:
return http.StatusForbidden
case KindUnavailable:
return http.StatusServiceUnavailable
case KindRateLimited:
return http.StatusTooManyRequests
default:
return http.StatusInternalServerError
}
}

// Constructors for common error kinds.

// NotFound creates a not-found error.
func NotFound(code, message string) *Error {
return &Error{Kind: KindNotFound, Code: code, Message: message}
}

// Invalid creates a validation error.
func Invalid(code, message string) *Error {
return &Error{Kind: KindInvalid, Code: code, Message: message}
}

// Conflict creates a conflict error (e.g., duplicate resource).
func Conflict(code, message string) *Error {
return &Error{Kind: KindConflict, Code: code, Message: message}
}

// Unauthorized creates an authentication error.
func Unauthorized(code, message string) *Error {
return &Error{Kind: KindUnauthorized, Code: code, Message: message}
}

// Forbidden creates an authorization error.
func Forbidden(code, message string) *Error {
return &Error{Kind: KindForbidden, Code: code, Message: message}
}

// Internal creates an internal error wrapping the cause.
func Internal(err error) *Error {
return &Error{
Kind:    KindInternal,
Code:    "internal_error",
Message: "internal error",
err:     err,
}
}

// WithDetails returns a copy of the error with validation details attached.
func (e *Error) WithDetails(details map[string]string) *Error {
cp := *e
cp.Details = details
return &cp
}

// IsKind reports whether err is an application error of the given kind.
func IsKind(err error, kind Kind) bool {
var appErr *Error
if errors.As(err, &appErr) {
return appErr.Kind == kind
}
return false
}
