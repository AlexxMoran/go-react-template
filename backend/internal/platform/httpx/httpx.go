// Package httpx holds transport-level helpers shared by every HTTP handler:
// JSON encode/decode with validation, the standard response envelopes
// (DataResponse / PaginatedResponse), error rendering, and pagination parsing.
//
// These mirror the Python backend's core/schemas/base.py envelopes and the
// centralized exception handlers.
package httpx

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/yourorg/goapp/pkg/apperror"
)

const (
	defaultLimit = 10
	maxLimit     = 100
)

// Validator is implemented by request DTOs that can self-validate. Valid returns
// a map of field -> message; an empty map means the input is valid.
type Validator interface {
	Valid(ctx context.Context) map[string]string
}

// DataResponse is the envelope for a single resource.
type DataResponse[T any] struct {
	Data T `json:"data"`
}

// PaginatedResponse is the envelope for a list of resources with paging metadata.
type PaginatedResponse[T any] struct {
	Data          []T   `json:"data"`
	Skip          int   `json:"skip"`
	Limit         int   `json:"limit"`
	FilteredCount int64 `json:"filtered_count"`
	TotalCount    int64 `json:"total_count"`
}

// Data writes a 1xx-2xx single-resource response.
func Data[T any](w http.ResponseWriter, status int, data T) {
	JSON(w, status, DataResponse[T]{Data: data})
}

// JSON serializes v as JSON with the given status code.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// Decode parses the JSON request body into T.
func Decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, apperror.BadRequest("invalid_body", "Request body is not valid JSON")
	}
	return v, nil
}

// DecodeValid parses the JSON body into T and runs its Valid method.
func DecodeValid[T Validator](r *http.Request) (T, error) {
	v, err := Decode[T](r)
	if err != nil {
		return v, err
	}
	if problems := v.Valid(r.Context()); len(problems) > 0 {
		return v, apperror.Validation("Validation failed", problems)
	}
	return v, nil
}

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Message    string            `json:"message"`
	MessageKey string            `json:"message_key"`
	Fields     map[string]string `json:"fields,omitempty"`
}

// WriteError renders err as a JSON error envelope. Non-apperror errors are
// treated as internal (500) and their detail is logged but never leaked.
func WriteError(w http.ResponseWriter, logger *slog.Logger, err error) {
	appErr, ok := apperror.As(err)
	if !ok {
		appErr = apperror.Internal(err)
	}
	if appErr.Status >= http.StatusInternalServerError && logger != nil {
		logger.Error("request failed", slog.String("message_key", appErr.MessageKey), slog.Any("err", appErr.Err))
	}
	JSON(w, appErr.Status, errorEnvelope{Error: errorBody{
		Message:    appErr.Message,
		MessageKey: appErr.MessageKey,
		Fields:     appErr.Fields,
	}})
}

// Pagination extracts skip/limit query parameters with sane bounds.
func Pagination(r *http.Request) (skip, limit int) {
	skip = atoiDefault(r.URL.Query().Get("skip"), 0)
	if skip < 0 {
		skip = 0
	}
	limit = atoiDefault(r.URL.Query().Get("limit"), defaultLimit)
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return skip, limit
}

func atoiDefault(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return n
}
