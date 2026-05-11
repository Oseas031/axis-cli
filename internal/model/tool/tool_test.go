package tool

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/axis-cli/axis/internal/types"
)

// mockTool is a simple tool implementation for testing the registry.
type mockTool struct {
	name string
}

func (m *mockTool) Name() string { return m.name }

func (m *mockTool) Schema() types.ToolDefinition {
	return types.ToolDefinition{
		Name:        m.name,
		Description: "mock tool for testing",
	}
}

func (m *mockTool) Execute(_ context.Context, _ map[string]any) (map[string]any, error) {
	return map[string]any{"result": "ok"}, nil
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	tool := &mockTool{name: "test-tool"}

	err := r.Register(tool, []string{"subprocess"})
	if err != nil {
		t.Fatalf("Register should succeed: %v", err)
	}

	got, ok := r.Get("test-tool")
	if !ok {
		t.Fatal("Get should return ok=true for registered tool")
	}
	if got.Name() != "test-tool" {
		t.Errorf("Expected name test-tool, got %s", got.Name())
	}
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	r := NewRegistry()
	tool := &mockTool{name: "dup"}

	err := r.Register(tool, []string{"subprocess"})
	if err != nil {
		t.Fatalf("First register should succeed: %v", err)
	}

	err = r.Register(&mockTool{name: "dup"}, []string{"subprocess"})
	if err == nil {
		t.Error("Duplicate register should return error")
	}
}

func TestRegistry_GetUnknown(t *testing.T) {
	r := NewRegistry()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Error("Get for unknown tool should return ok=false")
	}
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockTool{name: "a"}, []string{"subprocess"})
	r.Register(&mockTool{name: "b"}, []string{"filesystem:read"})

	defs := r.List()
	if len(defs) != 2 {
		t.Fatalf("Expected 2 tool definitions, got %d", len(defs))
	}

	names := make(map[string]bool)
	for _, d := range defs {
		names[d.Name] = true
	}
	if !names["a"] || !names["b"] {
		t.Error("List should contain both registered tool names")
	}
}

func TestRegistry_ListEmpty(t *testing.T) {
	r := NewRegistry()
	defs := r.List()
	if len(defs) != 0 {
		t.Errorf("Expected empty list, got %d items", len(defs))
	}
}

func TestRegistry_ImplementsInterface(t *testing.T) {
	var _ Tool = &mockTool{}
}

func TestRegistry_GetScopes(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockTool{name: "tool1"}, []string{"filesystem:read", "filesystem:write"})
	r.Register(&mockTool{name: "tool2"}, []string{"network"})

	scopes := r.GetScopes("tool1")
	if len(scopes) != 2 {
		t.Fatalf("Expected 2 scopes, got %d", len(scopes))
	}

	scopes2 := r.GetScopes("tool2")
	if len(scopes2) != 1 || scopes2[0] != "network" {
		t.Errorf("Expected [network], got %v", scopes2)
	}

	scopes3 := r.GetScopes("nonexistent")
	if scopes3 != nil {
		t.Errorf("Expected nil for unknown tool, got %v", scopes3)
	}
}

func TestToolPermissionScope_Constants(t *testing.T) {
	if ScopeFilesystemRead != "filesystem:read" {
		t.Errorf("Expected filesystem:read, got %s", ScopeFilesystemRead)
	}
	if ScopeFilesystemWrite != "filesystem:write" {
		t.Errorf("Expected filesystem:write, got %s", ScopeFilesystemWrite)
	}
	if ScopeNetwork != "network" {
		t.Errorf("Expected network, got %s", ScopeNetwork)
	}
	if ScopeSubprocess != "subprocess" {
		t.Errorf("Expected subprocess, got %s", ScopeSubprocess)
	}
}

// FileReadTool tests

func TestFileReadTool_Name(t *testing.T) {
	tool := NewFileReadTool([]string{"/tmp"})
	if tool.Name() != "file_read" {
		t.Errorf("Expected file_read, got %s", tool.Name())
	}
}

func TestFileReadTool_Schema(t *testing.T) {
	tool := NewFileReadTool([]string{"/tmp"})
	schema := tool.Schema()
	if schema.Name != "file_read" {
		t.Errorf("Expected name file_read, got %s", schema.Name)
	}
	if len(schema.Parameters) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(schema.Parameters))
	}
	if schema.Parameters[0].Name != "path" {
		t.Errorf("Expected parameter name path, got %s", schema.Parameters[0].Name)
	}
}

func TestFileReadTool_Execute_Success(t *testing.T) {
	// Create temp file in temp dir
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("hello world"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tool := NewFileReadTool([]string{tmpDir})
	result, err := tool.Execute(context.Background(), map[string]any{"path": testFile})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result["error"] != nil {
		t.Errorf("Expected no error, got %v", result["error"])
	}
	if result["content"] != "hello world" {
		t.Errorf("Expected 'hello world', got %v", result["content"])
	}
}

func TestFileReadTool_Execute_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewFileReadTool([]string{tmpDir})

	// Attempt path traversal
	result, err := tool.Execute(context.Background(), map[string]any{"path": tmpDir + "/../../../etc/passwd"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result["error"] == nil {
		t.Error("Expected error for path traversal attack")
	}
}

func TestFileReadTool_Execute_MissingPath(t *testing.T) {
	tool := NewFileReadTool([]string{"/tmp"})
	result, err := tool.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result["error"] == nil {
		t.Error("Expected error for missing path")
	}
}

func TestFileReadTool_Execute_OutsideAllowedDir(t *testing.T) {
	tmpDir := t.TempDir()
	otherDir := t.TempDir()

	tool := NewFileReadTool([]string{tmpDir})
	result, err := tool.Execute(context.Background(), map[string]any{"path": otherDir + "/file.txt"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result["error"] == nil {
		t.Error("Expected error for path outside allowed dir")
	}
}

// FileWriteTool tests

func TestFileWriteTool_Name(t *testing.T) {
	tool := NewFileWriteTool([]string{"/tmp"})
	if tool.Name() != "file_write" {
		t.Errorf("Expected file_write, got %s", tool.Name())
	}
}

func TestFileWriteTool_Schema(t *testing.T) {
	tool := NewFileWriteTool([]string{"/tmp"})
	schema := tool.Schema()
	if schema.Name != "file_write" {
		t.Errorf("Expected name file_write, got %s", schema.Name)
	}
	if len(schema.Parameters) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(schema.Parameters))
	}
}

func TestFileWriteTool_Execute_Success(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewFileWriteTool([]string{tmpDir})

	result, err := tool.Execute(context.Background(), map[string]any{
		"path":    filepath.Join(tmpDir, "output.txt"),
		"content": "hello write",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result["error"] != nil {
		t.Errorf("Expected no error, got %v", result["error"])
	}
	if result["success"] != true {
		t.Errorf("Expected success=true, got %v", result["success"])
	}

	// Verify file was written
	content, err := os.ReadFile(filepath.Join(tmpDir, "output.txt"))
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}
	if string(content) != "hello write" {
		t.Errorf("Expected 'hello write', got %s", string(content))
	}
}

func TestFileWriteTool_Execute_CreatesParentDirs(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewFileWriteTool([]string{tmpDir})

	nestedPath := filepath.Join(tmpDir, "a", "b", "c", "nested.txt")
	result, err := tool.Execute(context.Background(), map[string]any{
		"path":    nestedPath,
		"content": "nested content",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result["error"] != nil {
		t.Errorf("Expected no error, got %v", result["error"])
	}

	content, err := os.ReadFile(nestedPath)
	if err != nil {
		t.Fatalf("Failed to read nested file: %v", err)
	}
	if string(content) != "nested content" {
		t.Errorf("Expected 'nested content', got %s", string(content))
	}
}

func TestFileWriteTool_Execute_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewFileWriteTool([]string{tmpDir})

	result, err := tool.Execute(context.Background(), map[string]any{
		"path":    tmpDir + "/../../../etc/test.txt",
		"content": "malicious",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result["error"] == nil {
		t.Error("Expected error for path traversal attack")
	}
}

func TestFileWriteTool_Execute_MissingPath(t *testing.T) {
	tool := NewFileWriteTool([]string{"/tmp"})
	result, err := tool.Execute(context.Background(), map[string]any{"content": "test"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result["error"] == nil {
		t.Error("Expected error for missing path")
	}
}

func TestFileWriteTool_Execute_MissingContent(t *testing.T) {
	tool := NewFileWriteTool([]string{"/tmp"})
	result, err := tool.Execute(context.Background(), map[string]any{"path": "/tmp/test.txt"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result["error"] == nil {
		t.Error("Expected error for missing content")
	}
}

// HTTPClientTool tests

func TestHTTPClientTool_Name(t *testing.T) {
	tool := NewHTTPClientTool([]string{"example.com"})
	if tool.Name() != "http_request" {
		t.Errorf("Expected http_request, got %s", tool.Name())
	}
}

func TestHTTPClientTool_Schema(t *testing.T) {
	tool := NewHTTPClientTool([]string{"example.com"})
	schema := tool.Schema()
	if schema.Name != "http_request" {
		t.Errorf("Expected name http_request, got %s", schema.Name)
	}
	if len(schema.Parameters) != 4 {
		t.Errorf("Expected 4 parameters, got %d", len(schema.Parameters))
	}
}

func TestHTTPClientTool_ValidateHost_Success(t *testing.T) {
	tool := NewHTTPClientTool([]string{"example.com", "api.example.org"})
	if err := tool.validateHost("https://example.com/path"); err != nil {
		t.Errorf("Expected no error for allowed host, got %v", err)
	}
	if err := tool.validateHost("http://api.example.org:8080"); err != nil {
		t.Errorf("Expected no error for allowed host with port, got %v", err)
	}
}

func TestHTTPClientTool_ValidateHost_Rejected(t *testing.T) {
	tool := NewHTTPClientTool([]string{"example.com"})
	if err := tool.validateHost("https://evil.com/path"); err == nil {
		t.Error("Expected error for disallowed host")
	}
}

func TestHTTPClientTool_ValidateHost_InvalidURL(t *testing.T) {
	tool := NewHTTPClientTool([]string{"example.com"})
	if err := tool.validateHost("not-a-valid-url"); err == nil {
		t.Error("Expected error for invalid URL")
	}
}

// PathValidationError tests

func TestPathValidationError_Error(t *testing.T) {
	err := &PathValidationError{Path: "/test", Reason: "out of bounds"}
	expected := "path validation failed: out of bounds: /test"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestHTTPClientTool_Execute_GetSuccess(t *testing.T) {
	serverURL, allowedHost := newHTTPToolTestServer(t)
	tool := NewHTTPClientTool([]string{allowedHost})
	result, err := tool.Execute(context.Background(), map[string]any{
		"method": "GET",
		"url":    serverURL + "/get",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result["error"] != nil {
		t.Errorf("Expected no error, got %v", result["error"])
	}
	if result["status"] == nil {
		t.Error("Expected status in result")
	}
}

func TestHTTPClientTool_Execute_PostSuccess(t *testing.T) {
	serverURL, allowedHost := newHTTPToolTestServer(t)
	tool := NewHTTPClientTool([]string{allowedHost})
	result, err := tool.Execute(context.Background(), map[string]any{
		"method": "POST",
		"url":    serverURL + "/post",
		"body":   `{"test":"data"}`,
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result["error"] != nil {
		t.Errorf("Expected no error, got %v", result["error"])
	}
	if result["status"] == nil {
		t.Error("Expected status in result")
	}
}

func TestHTTPClientTool_Execute_DisallowedHost(t *testing.T) {
	tool := NewHTTPClientTool([]string{"example.com"})
	result, err := tool.Execute(context.Background(), map[string]any{
		"method": "GET",
		"url":    "https://evil.com/",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result["error"] == nil {
		t.Error("Expected error for disallowed host")
	}
}

func TestHTTPClientTool_Execute_MissingMethod(t *testing.T) {
	tool := NewHTTPClientTool([]string{"example.com"})
	result, err := tool.Execute(context.Background(), map[string]any{
		"url": "https://example.com",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result["error"] == nil {
		t.Error("Expected error for missing method")
	}
}

func TestHTTPClientTool_Execute_MissingURL(t *testing.T) {
	tool := NewHTTPClientTool([]string{"example.com"})
	result, err := tool.Execute(context.Background(), map[string]any{
		"method": "GET",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result["error"] == nil {
		t.Error("Expected error for missing URL")
	}
}

func TestHTTPClientTool_Execute_WithHeaders(t *testing.T) {
	serverURL, allowedHost := newHTTPToolTestServer(t)
	tool := NewHTTPClientTool([]string{allowedHost})
	result, err := tool.Execute(context.Background(), map[string]any{
		"method": "GET",
		"url":    serverURL + "/get",
		"headers": map[string]any{
			"X-Test-Header": "test-value",
		},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result["error"] != nil {
		t.Errorf("Expected no error, got %v", result["error"])
	}
}

func TestHTTPClientTool_Execute_PutDelete(t *testing.T) {
	serverURL, allowedHost := newHTTPToolTestServer(t)
	tool := NewHTTPClientTool([]string{allowedHost})

	// Test PUT
	result, err := tool.Execute(context.Background(), map[string]any{
		"method": "PUT",
		"url":    serverURL + "/put",
		"body":   `{"update":true}`,
	})
	if err != nil {
		t.Fatalf("Execute PUT failed: %v", err)
	}
	if result["error"] != nil {
		t.Errorf("Expected no error for PUT, got %v", result["error"])
	}

	// Test DELETE
	result, err = tool.Execute(context.Background(), map[string]any{
		"method": "DELETE",
		"url":    serverURL + "/delete",
	})
	if err != nil {
		t.Fatalf("Execute DELETE failed: %v", err)
	}
	if result["error"] != nil {
		t.Errorf("Expected no error for DELETE, got %v", result["error"])
	}
}

func TestHostValidationError_Error(t *testing.T) {
	err := &HostValidationError{URL: "http://evil.com", Reason: "not allowed"}
	expected := "host validation failed: not allowed: http://evil.com"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func newHTTPToolTestServer(t *testing.T) (string, string) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"method": r.Method,
			"path":   r.URL.Path,
			"header": r.Header.Get("X-Test-Header"),
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	parsed, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("failed to parse test server URL: %v", err)
	}
	return server.URL, parsed.Hostname()
}
