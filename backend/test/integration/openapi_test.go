//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"

	"github.com/yourorg/goapp/api"
	"github.com/yourorg/goapp/test/testsupport"
)

func loadSpec(t *testing.T) *openapi3.T {
	t.Helper()
	doc, err := openapi3.NewLoader().LoadFromData(api.Spec)
	if err != nil {
		t.Fatalf("load spec: %v", err)
	}
	return doc
}

// TestOpenAPI_SpecIsValid checks the embedded document is a well-formed OpenAPI 3
// spec (catches typos/broken $refs at build time).
func TestOpenAPI_SpecIsValid(t *testing.T) {
	if err := loadSpec(t).Validate(context.Background()); err != nil {
		t.Fatalf("openapi spec is invalid: %v", err)
	}
}

// TestOpenAPI_ResponsesMatchSpec drives the real server and validates each
// response against the spec. This is the drift guard: if a handler returns a
// shape the spec doesn't describe (or vice versa), the test fails.
func TestOpenAPI_ResponsesMatchSpec(t *testing.T) {
	testsupport.Truncate(t, pool)

	doc := loadSpec(t)
	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		t.Fatalf("build router: %v", err)
	}
	srv := newTestServer()

	// validate runs a request through the server and checks the response body
	// against the spec, returning the recorder for further assertions.
	validate := func(method, path, token string, body any) *httptest.ResponseRecorder {
		t.Helper()
		var reader io.Reader
		if body != nil {
			b, _ := json.Marshal(body)
			reader = bytes.NewReader(b)
		}
		req := httptest.NewRequestWithContext(context.Background(), method, path, reader)
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		route, pathParams, err := router.FindRoute(req)
		if err != nil {
			t.Fatalf("find route %s %s: %v", method, path, err)
		}
		input := &openapi3filter.ResponseValidationInput{
			RequestValidationInput: &openapi3filter.RequestValidationInput{
				Request:    req,
				PathParams: pathParams,
				Route:      route,
				Options:    &openapi3filter.Options{AuthenticationFunc: openapi3filter.NoopAuthenticationFunc},
			},
			Status: rec.Code,
			Header: rec.Header(),
		}
		input.SetBodyBytes(rec.Body.Bytes())
		if err := openapi3filter.ValidateResponse(context.Background(), input); err != nil {
			t.Errorf("%s %s -> %d response violates spec: %v\nbody: %s",
				method, path, rec.Code, err, rec.Body.String())
		}
		return rec
	}

	validate(http.MethodPost, "/api/auth/register", "",
		map[string]string{"email": "spec@example.com", "password": "password123", "first_name": "Spec"})

	loginRec := validate(http.MethodPost, "/api/auth/login", "",
		map[string]string{"email": "spec@example.com", "password": "password123"})
	var tok struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(loginRec.Body.Bytes(), &tok); err != nil || tok.AccessToken == "" {
		t.Fatalf("login token: %v (body %s)", err, loginRec.Body.String())
	}

	validate(http.MethodGet, "/api/auth/me", tok.AccessToken, nil)
	validate(http.MethodGet, "/api/v1/articles/", "", nil)

	createRec := validate(http.MethodPost, "/api/v1/articles/", tok.AccessToken,
		map[string]string{"title": "Spec article", "content": "body"})
	var created struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil || created.Data.ID == 0 {
		t.Fatalf("create article: %v (body %s)", err, createRec.Body.String())
	}

	validate(http.MethodGet, fmt.Sprintf("/api/v1/articles/%d", created.Data.ID), tok.AccessToken, nil)
	validate(http.MethodGet, "/api/v1/articles/999999", "", nil) // 404 error envelope
}
