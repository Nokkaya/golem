package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFileTool(t *testing.T) {
	tmpDir := t.TempDir()
	tool, err := NewWriteFileTool(tmpDir)
	if err != nil {
		t.Fatalf("NewWriteFileTool error: %v", err)
	}

	targetFile := filepath.Join(tmpDir, "output.txt")
	content := "hello world\nsecond line"
	argsJSON := fmt.Sprintf(`{"path": %q, "content": %q}`, targetFile, content)

	ctx := context.Background()
	result, err := tool.InvokableRun(ctx, argsJSON)
	if err != nil {
		t.Fatalf("InvokableRun error: %v", err)
	}

	if !strings.Contains(result, "success") && !strings.Contains(result, "Success") && !strings.Contains(result, "successfully") {
		t.Errorf("expected success message, got: %s", result)
	}

	data, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if string(data) != content {
		t.Errorf("expected file content %q, got %q", content, string(data))
	}
}

func TestListDirTool(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some files and a subdirectory
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("b"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	tool, err := NewListDirTool(tmpDir)
	if err != nil {
		t.Fatalf("NewListDirTool error: %v", err)
	}

	ctx := context.Background()
	argsJSON := fmt.Sprintf(`{"path": %q}`, tmpDir)

	result, err := tool.InvokableRun(ctx, argsJSON)
	if err != nil {
		t.Fatalf("InvokableRun error: %v", err)
	}

	// The result should be a JSON array of strings
	if !strings.Contains(result, "file1.txt") {
		t.Errorf("expected result to contain 'file1.txt', got: %s", result)
	}
	if !strings.Contains(result, "file2.txt") {
		t.Errorf("expected result to contain 'file2.txt', got: %s", result)
	}
	if !strings.Contains(result, "subdir/") {
		t.Errorf("expected result to contain 'subdir/' (with trailing slash), got: %s", result)
	}
}

func TestReadFile_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	tool, err := NewReadFileTool(tmpDir)
	if err != nil {
		t.Fatalf("NewReadFileTool error: %v", err)
	}

	// Attempt to read a path outside workspace using .. traversal
	outsidePath := filepath.Join(tmpDir, "..", "outside.txt")
	argsJSON := fmt.Sprintf(`{"path": %q}`, outsidePath)

	ctx := context.Background()
	_, err = tool.InvokableRun(ctx, argsJSON)
	if err == nil {
		t.Fatal("expected error for path traversal, got nil")
	}
	if !strings.Contains(err.Error(), "access denied") && !strings.Contains(err.Error(), "outside workspace") {
		t.Errorf("expected access denied error, got: %v", err)
	}
}

func TestWriteFile_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	tool, err := NewWriteFileTool(tmpDir)
	if err != nil {
		t.Fatalf("NewWriteFileTool error: %v", err)
	}

	outsidePath := filepath.Join(tmpDir, "..", "evil.txt")
	argsJSON := fmt.Sprintf(`{"path": %q, "content": "malicious"}`, outsidePath)

	ctx := context.Background()
	_, err = tool.InvokableRun(ctx, argsJSON)
	if err == nil {
		t.Fatal("expected error for path traversal, got nil")
	}
	if !strings.Contains(err.Error(), "access denied") && !strings.Contains(err.Error(), "outside workspace") {
		t.Errorf("expected access denied error, got: %v", err)
	}
}

func TestListDir_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	tool, err := NewListDirTool(tmpDir)
	if err != nil {
		t.Fatalf("NewListDirTool error: %v", err)
	}

	outsidePath := filepath.Join(tmpDir, "..")
	argsJSON := fmt.Sprintf(`{"path": %q}`, outsidePath)

	ctx := context.Background()
	_, err = tool.InvokableRun(ctx, argsJSON)
	if err == nil {
		t.Fatal("expected error for path traversal, got nil")
	}
	if !strings.Contains(err.Error(), "access denied") && !strings.Contains(err.Error(), "outside workspace") {
		t.Errorf("expected access denied error, got: %v", err)
	}
}

func TestReadFile_EmptyWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "readable.txt")
	os.WriteFile(testFile, []byte("content here"), 0644)

	// Empty workspace means no path restriction (backward compat)
	tool, err := NewReadFileTool("")
	if err != nil {
		t.Fatalf("NewReadFileTool error: %v", err)
	}

	ctx := context.Background()
	argsJSON := fmt.Sprintf(`{"path": %q}`, testFile)

	result, err := tool.InvokableRun(ctx, argsJSON)
	if err != nil {
		t.Fatalf("InvokableRun error: %v", err)
	}

	// Parse result to verify content
	var output ReadFileOutput
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		// If it doesn't parse as ReadFileOutput, just check raw string
		if !strings.Contains(result, "content here") {
			t.Errorf("expected result to contain 'content here', got: %s", result)
		}
		return
	}
	if !strings.Contains(output.Content, "content here") {
		t.Errorf("expected content to contain 'content here', got: %s", output.Content)
	}
}

func TestReadFile_SymlinkEscapeBlocked(t *testing.T) {
	workspace := t.TempDir()
	outside := t.TempDir()
	secretFile := filepath.Join(outside, "secret.txt")
	if err := os.WriteFile(secretFile, []byte("top-secret"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	linkPath := filepath.Join(workspace, "link-out")
	if err := os.Symlink(outside, linkPath); err != nil {
		t.Skipf("symlink not supported on this environment: %v", err)
	}

	tool, err := NewReadFileTool(workspace)
	if err != nil {
		t.Fatalf("NewReadFileTool error: %v", err)
	}

	argsJSON := fmt.Sprintf(`{"path": %q}`, filepath.Join(linkPath, "secret.txt"))
	_, err = tool.InvokableRun(context.Background(), argsJSON)
	if err == nil {
		t.Fatal("expected error for symlink escape, got nil")
	}
	if !strings.Contains(err.Error(), "access denied") {
		t.Fatalf("expected access denied, got: %v", err)
	}
}

func TestWriteFile_SymlinkEscapeBlocked(t *testing.T) {
	workspace := t.TempDir()
	outside := t.TempDir()

	linkPath := filepath.Join(workspace, "link-out")
	if err := os.Symlink(outside, linkPath); err != nil {
		t.Skipf("symlink not supported on this environment: %v", err)
	}

	tool, err := NewWriteFileTool(workspace)
	if err != nil {
		t.Fatalf("NewWriteFileTool error: %v", err)
	}

	target := filepath.Join(linkPath, "evil.txt")
	argsJSON := fmt.Sprintf(`{"path": %q, "content": "malicious"}`, target)
	_, err = tool.InvokableRun(context.Background(), argsJSON)
	if err == nil {
		t.Fatal("expected error for symlink escape, got nil")
	}
	if !strings.Contains(err.Error(), "access denied") {
		t.Fatalf("expected access denied, got: %v", err)
	}
}

func TestPathTraversal_ReadFile(t *testing.T) {
	tmpDir := t.TempDir()
	outsideFile := filepath.Join(tmpDir, "outside.txt")
	if err := os.WriteFile(outsideFile, []byte("secret content"), 0644); err != nil {
		t.Fatalf("failed to create outside file: %v", err)
	}

	workspaceDir := filepath.Join(tmpDir, "workspace")
	if err := os.Mkdir(workspaceDir, 0755); err != nil {
		t.Fatalf("failed to create workspace dir: %v", err)
	}

	tool, err := NewReadFileTool(workspaceDir)
	if err != nil {
		t.Fatalf("NewReadFileTool error: %v", err)
	}

	argsJSON := fmt.Sprintf(`{"path": %q}`, outsideFile)
	result, err := tool.InvokableRun(context.Background(), argsJSON)
	if err == nil {
		t.Fatalf("SECURITY FAILURE: Was able to read file outside workspace! Result: %s", result)
	}
	if !strings.Contains(err.Error(), "access denied") {
		t.Errorf("Expected 'access denied' error, got: %v", err)
	}
}

func TestPathTraversal_RelativePath(t *testing.T) {
	tmpDir := t.TempDir()
	outsideFile := filepath.Join(tmpDir, "outside.txt")
	if err := os.WriteFile(outsideFile, []byte("secret content"), 0644); err != nil {
		t.Fatalf("failed to create outside file: %v", err)
	}

	workspaceDir := filepath.Join(tmpDir, "workspace")
	if err := os.Mkdir(workspaceDir, 0755); err != nil {
		t.Fatalf("failed to create workspace dir: %v", err)
	}

	tool, err := NewReadFileTool(workspaceDir)
	if err != nil {
		t.Fatalf("NewReadFileTool error: %v", err)
	}

	result, err := tool.InvokableRun(context.Background(), `{"path": "../outside.txt"}`)
	if err == nil {
		t.Fatalf("SECURITY FAILURE: Was able to read file outside workspace via relative path! Result: %s", result)
	}
	if !strings.Contains(err.Error(), "access denied") {
		t.Errorf("Expected 'access denied' error, got: %v", err)
	}
}

func TestReadFileEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	tool, err := NewReadFileTool(tmpDir)
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}
	ctx := context.Background()

	tests := []struct {
		name     string
		content  string
		offset   int
		limit    int
		expected string
		lines    int
	}{
		{name: "Empty file", content: "", offset: 0, limit: 0, expected: "", lines: 1},
		{name: "One line no newline", content: "hello", offset: 0, limit: 0, expected: "hello", lines: 1},
		{name: "One line with newline", content: "hello\n", offset: 0, limit: 0, expected: "hello\n", lines: 2},
		{name: "Two lines", content: "hello\nworld", offset: 0, limit: 0, expected: "hello\nworld", lines: 2},
		{name: "Two lines with trailing newline", content: "hello\nworld\n", offset: 0, limit: 0, expected: "hello\nworld\n", lines: 3},
		{name: "Offset 1", content: "line1\nline2\nline3", offset: 1, limit: 1, expected: "line2", lines: 3},
		{name: "Offset 1 Limit 0 (rest of file)", content: "line1\nline2\nline3", offset: 1, limit: 0, expected: "line2\nline3", lines: 3},
		{name: "Offset out of bounds", content: "line1", offset: 5, limit: 1, expected: "", lines: 1},
		{name: "Limit spans newline", content: "line1\nline2\nline3", offset: 0, limit: 2, expected: "line1\nline2", lines: 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fname := filepath.Join(tmpDir, tc.name+".txt")
			if err := os.WriteFile(fname, []byte(tc.content), 0644); err != nil {
				t.Fatal(err)
			}

			input := fmt.Sprintf(`{"path": "%s.txt", "offset": %d, "limit": %d}`, tc.name, tc.offset, tc.limit)
			outStr, err := tool.InvokableRun(ctx, input)
			if err != nil {
				t.Fatalf("run failed: %v", err)
			}

			var out ReadFileOutput
			if err := json.Unmarshal([]byte(outStr), &out); err != nil {
				t.Fatalf("failed to unmarshal output: %v", err)
			}
			if out.Content != tc.expected {
				t.Errorf("expected content %q, got %q", tc.expected, out.Content)
			}
			if out.TotalLines != tc.lines {
				t.Errorf("expected total lines %d, got %d", tc.lines, out.TotalLines)
			}
		})
	}
}
