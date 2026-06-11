// Package arch holds architecture guardrail tests. They parse import statements
// statically (no code execution, no database) and fail the build when a layer
// boundary is crossed — the Go counterpart of the Python
// tests/unit/test_architecture.py.
package arch

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// moduleRoot walks up from the test's working directory until it finds go.mod.
func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

func imports(t *testing.T, path string) []string {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	out := make([]string, 0, len(file.Imports))
	for _, imp := range file.Imports {
		out = append(out, strings.Trim(imp.Path.Value, `"`))
	}
	return out
}

func walkGoFiles(t *testing.T, root string, fn func(path string)) {
	t.Helper()
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || name == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".go") {
			fn(path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
}

// TestDecisionsArePure enforces that operation decision layers contain only pure
// business rules: no database driver, no generated DB code, and no HTTP/router
// packages. This keeps them unit-testable without a database — exactly the
// invariant guarded on the Python decisions layer.
func TestDecisionsArePure(t *testing.T) {
	root := moduleRoot(t)
	forbidden := []string{
		"github.com/jackc/pgx",
		"/database/gen",
		"net/http",
		"github.com/gin-gonic/gin",
	}

	checked := 0
	walkGoFiles(t, filepath.Join(root, "internal"), func(path string) {
		if filepath.Base(path) != "decisions.go" || !strings.Contains(filepath.ToSlash(path), "/operation/") {
			return
		}
		checked++
		for _, imp := range imports(t, path) {
			for _, bad := range forbidden {
				if strings.Contains(imp, bad) {
					rel, _ := filepath.Rel(root, path)
					t.Errorf("%s: decision layer must not import %q (keep it pure)", rel, imp)
				}
			}
		}
	})
	if checked == 0 {
		t.Fatal("no operation decisions.go files were checked")
	}
}

// TestGinStaysAtHTTPBoundary keeps the framework context from leaking into
// services, repositories, policies, and operation internals.
func TestGinStaysAtHTTPBoundary(t *testing.T) {
	root := moduleRoot(t)
	allowedFiles := map[string]bool{
		"server.go":     true,
		"handler.go":    true,
		"routes.go":     true,
		"middleware.go": true,
		"httpx.go":      true,
	}
	walkGoFiles(t, filepath.Join(root, "internal"), func(path string) {
		for _, imp := range imports(t, path) {
			if imp != "github.com/gin-gonic/gin" {
				continue
			}
			if allowedFiles[filepath.Base(path)] {
				continue
			}
			rel, _ := filepath.Rel(root, path)
			t.Errorf("%s: gin must stay at the HTTP boundary", rel)
		}
	})
}

// TestServicesDoNotAuthorize enforces that domain service files never perform
// authorization. Authorization belongs exclusively at the HTTP entrypoint
// (handlers), so services stay reusable from system-triggered callers. Mirrors
// the Python rule "module service files must not import core.permissions".
func TestServicesDoNotAuthorize(t *testing.T) {
	root := moduleRoot(t)
	walkGoFiles(t, filepath.Join(root, "internal"), func(path string) {
		if filepath.Base(path) != "service.go" {
			return
		}
		for _, imp := range imports(t, path) {
			if strings.HasSuffix(imp, "/platform/authz") {
				rel, _ := filepath.Rel(root, path)
				t.Errorf("%s: service must not import authz (authorize at the handler instead)", rel)
			}
		}
	})
}
