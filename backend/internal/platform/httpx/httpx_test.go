package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/yourorg/goapp/pkg/apperror"
)

func TestPaginationAppliesDefaultsAndBounds(t *testing.T) {
	c, _ := testContext("GET", "/items?skip=-10&limit=1000")

	skip, limit := Pagination(c)

	if skip != 0 {
		t.Fatalf("skip = %d, want 0", skip)
	}
	if limit != maxLimit {
		t.Fatalf("limit = %d, want maxLimit %d", limit, maxLimit)
	}
}

func TestWriteErrorUsesStandardEnvelope(t *testing.T) {
	c, rec := testContext("GET", "/items")

	WriteError(c, slog.Default(), apperror.Validation("Validation failed", map[string]string{
		"title": "is required",
	}))

	if c.Writer.Status() != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", c.Writer.Status())
	}

	var body struct {
		Error struct {
			Message    string            `json:"message"`
			MessageKey string            `json:"message_key"`
			Fields     map[string]string `json:"fields"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.MessageKey != "validation_error" || body.Error.Fields["title"] != "is required" {
		t.Fatalf("error body = %+v, want validation envelope", body.Error)
	}
}

func testContext(method, target string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(method, target, nil)
	return c, rec
}
