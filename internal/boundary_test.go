package internal_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func findRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("cannot find project root")
		}
		dir = parent
	}
}

func importsIn(t *testing.T, dir string) []string {
	t.Helper()
	var imports []string
	fset := token.NewFileSet()
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return nil // skip unparseable files
		}
		for _, imp := range f.Imports {
			imports = append(imports, strings.Trim(imp.Path.Value, `"`))
		}
		return nil
	})
	return imports
}

func TestBoundaryKernelSchedulerDispatcherNoModel(t *testing.T) {
	root := findRoot(t)
	dirs := []string{
		filepath.Join(root, "internal", "kernel", "scheduler"),
		filepath.Join(root, "internal", "kernel", "dispatcher"),
	}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}
		t.Run(filepath.Base(dir), func(t *testing.T) {
			for _, imp := range importsIn(t, dir) {
				if strings.Contains(imp, "internal/model") {
					t.Errorf("BOUNDARY violation: %s imports %q", filepath.Base(dir), imp)
				}
			}
		})
	}
}

func TestBoundaryMemoryNoExternalDeps(t *testing.T) {
	root := findRoot(t)
	dir := filepath.Join(root, "internal", "memory")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Skip("internal/memory not found")
	}
	for _, imp := range importsIn(t, dir) {
		if strings.Contains(imp, "github.com") && !strings.Contains(imp, "github.com/axis-cli/axis") {
			t.Errorf("BOUNDARY violation: memory imports external %q", imp)
		}
	}
}

func TestBoundaryContextpackIsolation(t *testing.T) {
	root := findRoot(t)
	dir := filepath.Join(root, "internal", "contextpack")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Skip("internal/contextpack not found")
	}
	for _, imp := range importsIn(t, dir) {
		if strings.Contains(imp, "internal/model") {
			t.Errorf("BOUNDARY violation: contextpack imports %q", imp)
		}
		if strings.Contains(imp, "internal/kernel") {
			t.Errorf("BOUNDARY violation: contextpack imports %q", imp)
		}
	}
}
