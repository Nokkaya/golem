package tools

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/cloudwego/eino/schema"
)

// Mock tool for testing
type mockTool struct{}

func (m *mockTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
    return &schema.ToolInfo{
        Name: "mock_tool",
        Desc: "A mock tool for testing",
    }, nil
}

func (m *mockTool) InvokableRun(ctx context.Context, args string, opts ...any) (string, error) {
    return "mock result", nil
}

func TestRegistry_RegisterAndGet(t *testing.T) {
    reg := NewRegistry()

    err := reg.Register(&mockTool{})
    if err != nil {
        t.Fatalf("Register error: %v", err)
    }

    tool, ok := reg.Get("mock_tool")
    if !ok {
        t.Fatal("expected to find mock_tool")
    }
    if tool == nil {
        t.Fatal("tool is nil")
    }
}

func TestReadFileTool(t *testing.T) {
    tmpDir := t.TempDir()
    testFile := filepath.Join(tmpDir, "test.txt")
    os.WriteFile(testFile, []byte("line1\nline2\nline3"), 0644)

    tool, err := NewReadFileTool()
    if err != nil {
        t.Fatalf("NewReadFileTool error: %v", err)
    }

    ctx := context.Background()
    argsJSON := fmt.Sprintf(`{"path": %q}`, testFile)

    result, err := tool.InvokableRun(ctx, argsJSON)
    if err != nil {
        t.Fatalf("InvokableRun error: %v", err)
    }

    if !strings.Contains(result, "line1") {
        t.Errorf("expected result to contain 'line1', got: %s", result)
    }
}
