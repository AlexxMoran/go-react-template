// Package apperror provides a single application error type carried across all
// layers. Every error that should reach the client as a controlled response is
// an *Error, mirroring the Python backend's AppException: it carries an HTTP
// status, a stable MessageKey for frontend i18n, a human-readable Message, and
// optional field-level details for validation errors.
package apperror

import (
	"errors"
	"net/http"
)

// Error is the canonical application error.
type Error struct {
	// Status is the HTTP status code to return.
	Status int
	// MessageKey is a stable, machine-readable key the frontend maps to a
	// localized string (e.g. "not_found", "forbidden", "invalid_credentials").
	MessageKey string
	// Message is a human-readable, English fallback message.
	Message string
	// Fields holds per-field validation messages, keyed by field name.
	Fields map[string]string
	// Err is an optional wrapped underlying error (never sent to the client).
	Err error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap exposes the wrapped error for errors.Is / errors.As.
func (e *Error) Unwrap() error { return e.Err }

// New builds an *Error with an explicit status, message key and message.
func New(status int, messageKey, message string) *Error {
	return &Error{Status: status, MessageKey: messageKey, Message: message}
}

// Wrap attaches an underlying error to a new *Error.
func Wrap(err error, status int, messageKey, message string) *Error {
	return &Error{Status: status, MessageKey: messageKey, Message: message, Err: err}
}

// As extracts an *Error from err's chain, if present.
func As(err error) (*Error, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

// ── Constructors for common cases ────────────────────────────────────────────

func NotFound(messageKey, message string) *Error {
	return New(http.StatusNotFound, messageKey, message)
}

func BadRequest(messageKey, message string) *Error {
	return New(http.StatusBadRequest, messageKey, message)
}

func Unauthorized(messageKey, message string) *Error {
	return New(http.StatusUnauthorized, messageKey, message)
}

func Forbidden(messageKey, message string) *Error {
	return New(http.StatusForbidden, messageKey, message)
}

func Conflict(messageKey, message string) *Error {
	return New(http.StatusConflict, messageKey, message)
}

// Validation builds a 422 error with per-field messages.
func Validation(message string, fields map[string]string) *Error {
	return &Error{
		Status:     http.StatusUnprocessableEntity,
		MessageKey: "validation_error",
		Message:    message,
		Fields:     fields,
	}
}

// Internal wraps an unexpected error as a 500 with a generic, safe message.
func Internal(err error) *Error {
	return &Error{
		Status:     http.StatusInternalServerError,
		MessageKey: "internal_error",
		Message:    "Internal server error",
		Err:        err,
	}
}
