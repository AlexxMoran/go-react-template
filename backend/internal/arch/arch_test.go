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

// TestDomainIsPure enforces that every module's domain package contains only
// pure business rules: no database driver, no generated DB code, and no
// HTTP/router packages. This keeps the model, status machine, policy and
// decisions unit-testable without a database — the invariant the synthesis
// architecture is built on.
func TestDomainIsPure(t *testing.T) {
	root := moduleRoot(t)
	forbidden := []string{
		"github.com/jackc/pgx",
		"/database/gen",
		"net/http",
		"github.com/gin-gonic/gin",
	}

	checked := 0
	walkGoFiles(t, filepath.Join(root, "internal"), func(path string) {
		if filepath.Base(filepath.Dir(path)) != "domain" {
			return
		}
		checked++
		for _, imp := range imports(t, path) {
			for _, bad := range forbidden {
				if strings.Contains(imp, bad) {
					rel, _ := filepath.Rel(root, path)
					t.Errorf("%s: domain package must not import %q (keep it pure)", rel, imp)
				}
			}
		}
	})
	if checked == 0 {
		t.Fatal("no domain package files were checked")
	}
}

// TestModuleBoundaries enforces that domain modules talk to each other only
// through their published `*api` package. A file in internal/<A>/ may import
// internal/<B>/... only when the import is internal/<B>/<B>api (or a subpackage
// of it). Platform packages are a shared kernel and stay importable everywhere;
// the composition root (internal/app) and these arch tests are exempt as
// sources because wiring modules together is precisely their job.
func TestModuleBoundaries(t *testing.T) {
	root := moduleRoot(t)
	const modulePrefix = "github.com/yourorg/goapp/internal/"
	exemptSource := map[string]bool{"platform": true, "arch": true, "app": true}

	internalDir := filepath.Join(root, "internal")
	checked := 0
	walkGoFiles(t, internalDir, func(path string) {
		ownModule := moduleOf(path, internalDir)
		if ownModule == "" || exemptSource[ownModule] {
			return
		}
		for _, imp := range imports(t, path) {
			if !strings.HasPrefix(imp, modulePrefix) {
				continue
			}
			segs := strings.Split(strings.TrimPrefix(imp, modulePrefix), "/")
			target := segs[0]
			if target == ownModule || target == "platform" {
				continue // same module or shared kernel
			}
			if len(segs) >= 2 && segs[1] == target+"api" {
				continue // the other module's published contract
			}
			rel, _ := filepath.Rel(root, path)
			t.Errorf("%s: module %q may reach module %q only via %s/%sapi, not %q",
				rel, ownModule, target, target, target, imp)
		}
		checked++
	})
	if checked == 0 {
		t.Fatal("no module files were checked")
	}
}

// moduleOf returns the module segment for a file under internalDir (e.g.
// internal/user/userapi/userapi.go -> "user"), or "" if the file sits directly
// in internal/.
func moduleOf(path, internalDir string) string {
	rel, err := filepath.Rel(internalDir, path)
	if err != nil {
		return ""
	}
	parts := strings.Split(filepath.ToSlash(rel), "/")
	if len(parts) < 2 {
		return ""
	}
	return parts[0]
}

// TestGinStaysAtHTTPBoundary keeps the framework context from leaking into
// services, repositories, policies, and operation internals. Gin is allowed at
// the HTTP boundary: the transport files (server/handler/routes/httpx) and the
// whole platform/middleware package, which exists precisely to hold cross-cutting
// HTTP middleware.
func TestGinStaysAtHTTPBoundary(t *testing.T) {
	root := moduleRoot(t)
	allowedFiles := map[string]bool{
		"server.go":     true,
		"handler.go":    true,
		"routes.go":     true,
		"httpx.go":      true,
		"httpx_test.go": true,
		"middleware.go": true, // domain-local HTTP middleware (auth, authz)
		"metrics.go":    true, // optional HTTP metrics endpoint/middleware
	}
	walkGoFiles(t, filepath.Join(root, "internal"), func(path string) {
		for _, imp := range imports(t, path) {
			if imp != "github.com/gin-gonic/gin" {
				continue
			}
			if allowedFiles[filepath.Base(path)] || filepath.Base(filepath.Dir(path)) == "middleware" {
				continue
			}
			rel, _ := filepath.Rel(root, path)
			t.Errorf("%s: gin must stay at the HTTP boundary", rel)
		}
	})
}

// TestAsynqStaysAtJobsBoundary keeps the asynq job broker out of domain and
// application code. asynq may be imported only by the platform/jobs runtime and
// by a module's jobs.go file (the producer/handler boundary) — the same
// discipline as gin at the HTTP boundary.
func TestAsynqStaysAtJobsBoundary(t *testing.T) {
	root := moduleRoot(t)
	found := false
	walkGoFiles(t, filepath.Join(root, "internal"), func(path string) {
		for _, imp := range imports(t, path) {
			if imp != "github.com/hibiken/asynq" {
				continue
			}
			found = true
			if filepath.Base(path) == "jobs.go" || filepath.Base(filepath.Dir(path)) == "jobs" {
				continue
			}
			rel, _ := filepath.Rel(root, path)
			t.Errorf("%s: asynq must stay at the jobs boundary (a jobs.go file or platform/jobs)", rel)
		}
	})
	if !found {
		t.Fatal("no asynq imports found — update TestAsynqStaysAtJobsBoundary")
	}
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
