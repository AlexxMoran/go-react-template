// Package httpx holds transport-level helpers shared by every HTTP handler:
// JSON encode/decode with validation, the standard response envelopes
// (DataResponse / PaginatedResponse), error rendering, and pagination parsing.
//
// These mirror the Python backend's core/schemas/base.py envelopes and the
// centralized exception handlers.
package httpx

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

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
func Data[T any](c *gin.Context, status int, data T) {
	JSON(c, status, DataResponse[T]{Data: data})
}

// JSON serializes v as JSON with the given status code.
func JSON(c *gin.Context, status int, v any) {
	c.JSON(status, v)
}

// Decode parses the JSON request body into T.
func Decode[T any](c *gin.Context) (T, error) {
	var v T
	if err := c.ShouldBindJSON(&v); err != nil {
		return v, apperror.BadRequest("invalid_body", "Request body is not valid JSON")
	}
	return v, nil
}

// DecodeValid parses the JSON body into T and runs its Valid method.
func DecodeValid[T Validator](c *gin.Context) (T, error) {
	v, err := Decode[T](c)
	if err != nil {
		return v, err
	}
	if problems := v.Valid(c.Request.Context()); len(problems) > 0 {
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
func WriteError(c *gin.Context, logger *slog.Logger, err error) {
	appErr, ok := apperror.As(err)
	if !ok {
		appErr = apperror.Internal(err)
	}
	if appErr.Status >= http.StatusInternalServerError && logger != nil {
		logger.Error("request failed", slog.String("message_key", appErr.MessageKey), slog.Any("err", appErr.Err))
	}
	JSON(c, appErr.Status, errorEnvelope{Error: errorBody{
		Message:    appErr.Message,
		MessageKey: appErr.MessageKey,
		Fields:     appErr.Fields,
	}})
}

// Pagination extracts skip/limit query parameters with sane bounds.
func Pagination(c *gin.Context) (skip, limit int) {
	skip = atoiDefault(c.Query("skip"), 0)
	if skip < 0 {
		skip = 0
	}
	limit = atoiDefault(c.Query("limit"), defaultLimit)
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
